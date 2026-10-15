package server

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

type frame struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type wsClient struct {
	t *testing.T
	c *websocket.Conn
}

func (e *testEnv) dial(t *testing.T, path string) *wsClient {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(e.ts.URL, "http")+path, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", path, err)
	}
	t.Cleanup(func() { c.CloseNow() })
	return &wsClient{t: t, c: c}
}

func (w *wsClient) send(s string) {
	w.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.c.Write(ctx, websocket.MessageText, []byte(s)); err != nil {
		w.t.Fatalf("write: %v", err)
	}
}

func (w *wsClient) read() frame {
	w.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, b, err := w.c.Read(ctx)
	if err != nil {
		w.t.Fatalf("read: %v", err)
	}
	var f frame
	if err := json.Unmarshal(b, &f); err != nil {
		w.t.Fatalf("frame %s: %v", b, err)
	}
	return f
}

// expect reads one frame and checks its type and that the payload contains sub.
func (w *wsClient) expect(typ, sub string) frame {
	w.t.Helper()
	f := w.read()
	if f.Type != typ || !strings.Contains(string(f.Payload), sub) {
		w.t.Fatalf("got %s %s, want %s containing %q", f.Type, f.Payload, typ, sub)
	}
	return f
}

// expectNothing asserts no frame arrives within a short window.
func (w *wsClient) expectNothing() {
	w.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	if _, b, err := w.c.Read(ctx); err == nil {
		w.t.Fatalf("unexpected frame %s", b)
	}
}

// join sends a join and waits until the server has bound the client (it acks
// joins only implicitly, so poll the roster).
func (e *testEnv) join(t *testing.T, w *wsClient, id string) {
	t.Helper()
	w.send(`{"type":"join","payload":{"client":"` + id + `"}}`)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, c := range e.srv.clients() {
			if c.ID == id {
				return
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("client %s never joined", id)
}

func TestSignaling_JoinRelayPINAndOnAir(t *testing.T) {
	e := newEnv(t, nil)
	alice := e.dial(t, "/ws")
	bob := e.dial(t, "/ws")
	e.join(t, alice, "alice")
	e.join(t, bob, "bob")

	// Offer from alice is relayed verbatim to bob only.
	alice.send(`{"type":"offer","payload":{"sdp":"v=0 alice"}}`)
	bob.expect("offer", `"sdp":"v=0 alice"`)
	bob.send(`{"type":"answer","payload":{"sdp":"v=0 bob"}}`)
	alice.expect("answer", `"sdp":"v=0 bob"`)
	alice.send(`{"type":"ice-candidate","payload":{"candidate":"cand:1"}}`)
	bob.expect("ice-candidate", "cand:1")

	// Unpaired clients cannot go on air.
	alice.send(`{"type":"switch-source","payload":{"target":"alice"}}`)
	alice.expect("error", "not paired")

	// PIN flow: wrong PIN errors, right PIN pairs (no ack frame).
	alice.send(`{"type":"submit-pin","payload":{"pin":"0000"}}`)
	alice.expect("error", "incorrect pin")
	alice.send(`{"type":"submit-pin","payload":{"pin":"1234"}}`)

	// Now switching broadcasts on-air-state to every client. (Sent on alice's
	// conn: frames on one conn are processed in order, so the PIN is applied first.)
	alice.send(`{"type":"switch-source","payload":{"target":"alice"}}`)
	alice.expect("on-air-state", `"onair":"alice","live":true`)
	bob.expect("on-air-state", `"onair":"alice","live":true`)

	_, body := e.do(t, "GET", "/api/status", nil)
	st := decode[Status](t, body)
	if st.OnAir == nil || *st.OnAir != "alice" || st.Clients != 2 {
		t.Fatalf("status = %s", body)
	}
	_, body = e.do(t, "GET", "/api/devices", nil)
	dr := decode[devicesResponse](t, body)
	if len(dr.Clients) != 2 || dr.Clients[0] != (Client{ID: "alice", OnAir: true}) || dr.Clients[1] != (Client{ID: "bob", OnAir: false}) {
		t.Fatalf("clients = %+v", dr.Clients)
	}

	// The on-air client leaving broadcasts nobody-on-air to the rest.
	alice.c.Close(websocket.StatusNormalClosure, "")
	bob.expect("on-air-state", `"live":false`)
	deadline := time.Now().Add(5 * time.Second)
	for len(e.srv.clients()) != 1 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if got := e.srv.clients(); len(got) != 1 || got[0].ID != "bob" {
		t.Fatalf("roster after leave = %+v", got)
	}
}

func TestSignaling_TransportErrors(t *testing.T) {
	e := newEnv(t, nil)
	a := e.dial(t, "/ws")

	a.send(`not json`)
	a.expect("error", "malformed frame")
	a.send(`{"type":"offer","payload":{"sdp":"x"}}`)
	a.expect("error", "join before")
	a.send(`{"type":"submit-pin","payload":{"pin":"1234"}}`)
	a.expect("error", "join before")
	a.send(`{"type":"join","payload":{"client":"  "}}`)
	a.expect("error", "missing client id")

	e.join(t, a, "dup")
	a.send(`{"type":"join","payload":{"client":"other"}}`)
	a.expect("error", "already joined")
	a.send(`{"type":"submit-pin","payload":{}}`)
	a.expect("error", "submit-pin")
	a.send(`{"type":"bogus","payload":{}}`)
	a.expect("error", "unknown message type")

	// A second conn claiming the same id is refused, and the error goes to
	// the claimant, not the owner.
	b := e.dial(t, "/ws")
	b.send(`{"type":"join","payload":{"client":"dup"}}`)
	b.expect("error", "already connected")
	a.expectNothing()
}

func TestSignaling_PINLockout(t *testing.T) {
	e := newEnv(t, nil) // attempt limit 3
	a := e.dial(t, "/ws")
	e.join(t, a, "mallory")
	for i := 0; i < 3; i++ {
		a.send(`{"type":"submit-pin","payload":{"pin":"9999"}}`)
		a.expect("error", "incorrect pin")
	}
	a.send(`{"type":"submit-pin","payload":{"pin":"1234"}}`)
	a.expect("error", "rejected")
}

func TestEvents_MetricsStatusAndLog(t *testing.T) {
	e := newEnv(t, nil)
	ev := e.dial(t, "/api/events")

	first := ev.readEvent()
	if first.Type != "status" {
		t.Fatalf("first event = %s", first.Type)
	}

	// Collect until we see a metrics frame (initial one plus 50ms ticks).
	sawMetrics := false
	for i := 0; i < 5 && !sawMetrics; i++ {
		raw := ev.readEvent()
		if _, err := time.Parse(time.RFC3339, raw.TS); err != nil {
			t.Fatalf("ts %q: %v", raw.TS, err)
		}
		if raw.Type == "metrics" {
			sawMetrics = true
			if m := decode[Metrics](t, raw.Data); m.RAMTotalMB != 4000 {
				t.Fatalf("metrics = %s", raw.Data)
			}
		}
	}
	if !sawMetrics {
		t.Fatal("no metrics frame")
	}

	// Running a pipeline pushes a log event and a status change.
	e.do(t, "POST", "/api/pipelines/default/run", nil)
	var sawLog, sawStatus bool
	for i := 0; i < 100 && !(sawLog && sawStatus); i++ {
		raw := ev.readEvent()
		switch raw.Type {
		case "log":
			if l := decode[LogData](t, raw.Data); l.Subsystem == "pipeline" && l.Level == "info" {
				sawLog = true
			}
		case "status":
			if st := decode[Status](t, raw.Data); st.Running != nil && *st.Running == "default" {
				sawStatus = true
			}
		}
	}
	if !sawLog || !sawStatus {
		t.Fatalf("log=%v status=%v after run", sawLog, sawStatus)
	}
}

type rawEvent struct {
	Type string          `json:"type"`
	TS   string          `json:"ts"`
	Data json.RawMessage `json:"data"`
}

func (w *wsClient) readEvent() rawEvent {
	w.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, b, err := w.c.Read(ctx)
	if err != nil {
		w.t.Fatalf("read: %v", err)
	}
	var ev rawEvent
	if err := json.Unmarshal(b, &ev); err != nil {
		w.t.Fatalf("event %s: %v", b, err)
	}
	return ev
}

func TestShutdown_ClosesSockets(t *testing.T) {
	e := newEnv(t, nil)
	ev := e.dial(t, "/api/events")
	sig := e.dial(t, "/ws")
	e.srv.Close()
	for _, c := range []*wsClient{ev, sig} {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		var err error
		for err == nil {
			_, _, err = c.c.Read(ctx)
		}
		cancel()
		if websocket.CloseStatus(err) != websocket.StatusGoingAway {
			t.Fatalf("close err = %v, want going away", err)
		}
	}
}
