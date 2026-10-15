package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// broker fans events out to /api/events subscribers. Sends never block: a
// subscriber whose buffer is full misses that event.
type broker struct {
	mu   sync.Mutex
	subs map[chan Event]struct{}
}

func newBroker() *broker { return &broker{subs: make(map[chan Event]struct{})} }

func (b *broker) subscribe() chan Event {
	ch := make(chan Event, 32)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *broker) unsubscribe(ch chan Event) {
	b.mu.Lock()
	delete(b.subs, ch)
	b.mu.Unlock()
}

func (b *broker) publish(ev Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}

func newEvent(typ string, data any) Event {
	return Event{Type: typ, TS: time.Now().UTC().Format(time.RFC3339), Data: data}
}

// logEvent publishes a "log" event and mirrors it to the server logger.
func (s *Server) logEvent(level, subsystem, msg string) {
	if s.logger != nil {
		s.logger.Printf("[%s] %s: %s", level, subsystem, msg)
	}
	s.events.publish(newEvent("log", LogData{Level: level, Subsystem: subsystem, Msg: msg}))
}

// handleEvents serves GET /api/events: an initial status + metrics frame, then
// metrics every MetricsInterval and status / log frames as they happen.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, s.acceptOptions())
	if err != nil {
		return // Accept already wrote the HTTP error
	}
	defer c.CloseNow()

	ch := s.events.subscribe()
	defer s.events.unsubscribe(ch)

	// The UI never sends on this socket; CloseRead drains control frames and
	// cancels ctx when the peer goes away.
	ctx := c.CloseRead(r.Context())

	write := func(ev Event) error {
		b, err := json.Marshal(ev)
		if err != nil {
			return err
		}
		wctx, cancel := context.WithTimeout(ctx, writeTimeout)
		defer cancel()
		return c.Write(wctx, websocket.MessageText, b)
	}

	st := s.Status()
	if write(newEvent("status", st)) != nil || write(newEvent("metrics", st.Metrics)) != nil {
		return
	}

	t := time.NewTicker(s.opts.MetricsInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			c.Close(websocket.StatusGoingAway, "server shutting down")
			return
		case <-t.C:
			if write(newEvent("metrics", s.metrics.Sample())) != nil {
				return
			}
		case ev := <-ch:
			if write(ev) != nil {
				return
			}
		}
	}
}
