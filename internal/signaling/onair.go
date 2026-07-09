package signaling

import "github.com/adeelahmad/gadgetcast/internal/core"

// OnAirStateFrom projects an on-air-state message from the current on-air id and
// whether the session is live. A zero id with live=false means nobody is on air.
func OnAirStateFrom(id core.ClientID, live bool) OnAirStateMessage {
	return OnAirStateMessage{OnAir: id, Live: live}
}
