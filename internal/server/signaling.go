package server

import (
	"context"
	"encoding/json"
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

// sigConn is one /ws connection. id is zero until a join is accepted; it is
// read and written only under Server.hubMu.
type sigConn struct {
	id     core.ClientID
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
	sc := &sigConn{send: make(chan []byte, sendQueueDepth), cancel: cancel}

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
		s.deliverLocked(s.hub.Process(session.SubmitPIN{Client: sc.id, PIN: p.PIN}))

	case string(signaling.TypeJoin):
		logMsg = s.joinLocked(sc, data)

	default:
		if sc.id.IsZero() {
			s.replyError(sc, "join before sending "+strings.TrimSpace(env.Type))
			break
		}
		s.deliverLocked(s.hub.Process(session.Inbound{From: sc.id, Data: data}))
	}
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
	s.hubMu.Unlock()

	s.logEvent("info", "session", "client "+id.String()+" left")
	s.publishStatusIfChanged()
}
