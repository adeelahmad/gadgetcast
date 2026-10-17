package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/coder/websocket"

	"github.com/adeelahmad/gadgetcast/internal/core"
	"github.com/adeelahmad/gadgetcast/internal/session"
	"github.com/adeelahmad/gadgetcast/internal/signaling"
)

// typeSubmitPIN is the transport-level client→server frame carrying a PIN
// (docs/api.md). It is not part of the signaling codec: the transport handles
// it and feeds session.SubmitPIN to the hub.
const typeSubmitPIN = "submit-pin"

// typeMediaOffer / typeMediaAnswer carry a WebRTC session between a paired
// client and the Pi itself (docs/api.md). Like submit-pin they are handled
// by the transport, not the hub, which only relays between clients.
const (
	typeMediaOffer  = "media-offer"
	typeMediaAnswer = "media-answer"
)

// sigConn is one /ws connection. id is zero until a join is accepted; it is
// read and written only under Server.hubMu.
type sigConn struct {
	id     core.ClientID
	addr   string // remote host, for PIN rate limiting
	send   chan []byte
	cancel context.CancelFunc
	once   sync.Once
}

// enqueue queues a frame without blocking; a peer that cannot keep up is dropped.
func (c *sigConn) enqueue(b []byte) {
	select {
	case c.send <- b:
	default:
		c.once.Do(c.cancel)
	}
}

// handleSignaling serves GET /ws.
func (s *Server) handleSignaling(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, s.acceptOptions())
	if err != nil {
		return
	}
	defer c.CloseNow()
	c.SetReadLimit(maxFrameBytes)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	sc := &sigConn{send: make(chan []byte, sendQueueDepth), cancel: cancel, addr: remoteHost(r.RemoteAddr)}

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		for {
			select {
			case <-ctx.Done():
				return
			case b := <-sc.send:
				wctx, wcancel := context.WithTimeout(ctx, writeTimeout)
				err := c.Write(wctx, websocket.MessageText, b)
				wcancel()
				if err != nil {
					cancel()
					return
				}
			}
		}
	}()

	go func() {
		select {
		case <-s.done:
			c.Close(websocket.StatusGoingAway, "server shutting down")
			cancel()
		case <-ctx.Done():
		}
	}()

	for {
		typ, data, err := c.Read(ctx)
		if err != nil {
			break
		}
		if typ != websocket.MessageText {
			s.replyError(sc, "binary frames are not supported")
			continue
		}
		s.handleFrame(sc, data)
	}

	s.disconnect(sc)
	cancel()
	<-writerDone
}

func (s *Server) replyError(sc *sigConn, reason string) {
	b, err := signaling.Encode(signaling.ErrorMessage{Reason: reason})
	if err == nil {
		sc.enqueue(b)
	}
}

// deliverLocked encodes each effect and queues it on the addressed conn.
// Effects for clients with no live conn are dropped. Caller holds hubMu.
func (s *Server) deliverLocked(effects []session.Effect) {
	for _, e := range effects {
		conn := s.conns[e.To]
		if conn == nil {
			continue
		}
		b, err := signaling.Encode(e.Msg)
		if err != nil {
			continue
		}
		conn.enqueue(b)
	}
}

// handleFrame routes one inbound text frame from sc into the hub.
func (s *Server) handleFrame(sc *sigConn, data []byte) {
	var env struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		s.replyError(sc, "malformed frame: "+err.Error())
		return
	}

	var logMsg string
	s.hubMu.Lock()
	switch env.Type {
	case typeSubmitPIN:
		if sc.id.IsZero() {
			s.replyError(sc, "join before submitting a PIN")
			break
		}
		var p struct {
			PIN string `json:"pin"`
		}
		if err := json.Unmarshal(env.Payload, &p); err != nil || p.PIN == "" {
			s.replyError(sc, "submit-pin: payload must be {\"pin\":\"…\"}")
			break
		}
		if wait := s.pins.wait(sc.addr); wait > 0 {
			s.replyError(sc, fmt.Sprintf("too many wrong PINs from this device; try again in %d s", int(wait.Seconds()+0.999)))
			break
		}
		effects := s.hub.Process(session.SubmitPIN{Client: sc.id, PIN: p.PIN})
		if s.pm.IsPaired(sc.id) {
			s.pins.success(sc.addr)
		} else {
			s.pins.fail(sc.addr)
		}
		s.deliverLocked(effects)

	case string(signaling.TypeJoin):
		logMsg = s.joinLocked(sc, data)

	case typeMediaOffer:
		id := sc.id
		var p struct {
			SDP string `json:"sdp"`
		}
		switch {
		case id.IsZero():
			s.replyError(sc, "join before sending media-offer")
		case !s.pm.IsPaired(id):
			s.replyError(sc, "pair with the PIN before sending media")
		case s.opts.Media == nil:
			s.replyError(sc, "media is not available on this server")
		case json.Unmarshal(env.Payload, &p) != nil || p.SDP == "":
			s.replyError(sc, "media-offer: payload must be {\"sdp\":\"…\"}")
		default:
			// ICE gathering can take a moment: answer off the hub lock.
			go s.answerMedia(sc, id, p.SDP)
		}

	default:
		if sc.id.IsZero() {
			s.replyError(sc, "join before sending "+strings.TrimSpace(env.Type))
			break
		}
		s.deliverLocked(s.hub.Process(session.Inbound{From: sc.id, Data: data}))
	}
	s.syncOnAirLocked()
	s.hubMu.Unlock()

	if logMsg != "" {
		s.logEvent("info", "session", logMsg)
	}
	s.publishStatusIfChanged()
}

// joinLocked binds sc to the client id in a join frame once the hub accepts
// it. Duplicate ids and re-joins are rejected by the transport so an error is
// never misrouted to the conn that already owns the id. Caller holds hubMu.
func (s *Server) joinLocked(sc *sigConn, data []byte) (logMsg string) {
	if !sc.id.IsZero() {
		s.replyError(sc, "already joined as "+sc.id.String())
		return ""
	}
	msg, err := signaling.Decode(data)
	if err != nil {
		s.replyError(sc, err.Error())
		return ""
	}
	id := msg.(signaling.JoinMessage).Client
	if s.conns[id] != nil {
		s.replyError(sc, "client id "+id.String()+" is already connected")
		return ""
	}
	effects := s.hub.Process(session.Inbound{From: id, Data: data})
	for _, e := range effects {
		if e.To == id {
			if em, ok := e.Msg.(signaling.ErrorMessage); ok {
				s.replyError(sc, em.Reason)
				return ""
			}
		}
	}
	sc.id = id
	s.conns[id] = sc
	s.order = append(s.order, id)
	s.deliverLocked(effects)
	return "client " + id.String() + " joined"
}

// disconnect removes sc from the hub and routing table, delivering any
// resulting on-air broadcast to the remaining clients.
func (s *Server) disconnect(sc *sigConn) {
	s.hubMu.Lock()
	id := sc.id
	if id.IsZero() {
		s.hubMu.Unlock()
		return
	}
	effects := s.hub.Process(session.Disconnect{Client: id})
	delete(s.conns, id)
	for i, c := range s.order {
		if c == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	sc.id = core.ClientID{}
	s.deliverLocked(effects)
	s.syncOnAirLocked()
	s.hubMu.Unlock()
	if s.opts.Media != nil {
		s.opts.Media.Drop(id)
	}

	s.logEvent("info", "session", "client "+id.String()+" left")
	s.publishStatusIfChanged()
}

// answerMedia answers a client's media offer and sends the answer back.
func (s *Server) answerMedia(sc *sigConn, id core.ClientID, offer string) {
	answer, err := s.opts.Media.Offer(id, offer)
	if err != nil {
		s.replyError(sc, err.Error())
		return
	}
	s.hubMu.Lock()
	gone := s.conns[id] != sc
	s.hubMu.Unlock()
	if gone { // the client left while ICE was gathering
		s.opts.Media.Drop(id)
		return
	}
	b, err := json.Marshal(map[string]any{"type": typeMediaAnswer, "payload": map[string]string{"sdp": answer}})
	if err == nil {
		sc.enqueue(b)
	}
	s.logEvent("info", "session", "client "+id.String()+" connected media")
}

// syncOnAirLocked points the media bridge at the switcher's on-air client.
// Caller holds hubMu.
func (s *Server) syncOnAirLocked() {
	if s.opts.Media == nil {
		return
	}
	id, live := s.sw.OnAir()
	s.opts.Media.SetOnAir(id, live)
}

// remoteHost strips the port from an http.Request.RemoteAddr.
func remoteHost(addr string) string {
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	return addr
}
