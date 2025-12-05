package target

import (
	"infralog/tfplan"
	"time"
)

// Target defines the interface for notification targets.
type Target interface {
	Write(*Payload) error
}

// Payload contains the change data sent to targets.
type Payload struct {
	Plan     *tfplan.Plan `json:"plan"`
	Datetime time.Time    `json:"datetime"`
}

// NewPayload creates a new Payload with the current timestamp.
func NewPayload(plan *tfplan.Plan) *Payload {
	return &Payload{
		Plan:     plan,
		Datetime: time.Now().UTC(),
	}
}
