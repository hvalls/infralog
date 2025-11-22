package target

import (
	"infralog/config"
	"infralog/tfstate"
	"time"
)

// Target defines the interface for notification targets.
type Target interface {
	Write(*Payload) error
}

// Payload contains the change data and metadata sent to targets.
type Payload struct {
	Diffs    *tfstate.StateDiff `json:"diffs"`
	Metadata Metadata           `json:"metadata"`
}

// Metadata contains contextual information about the state change event.
type Metadata struct {
	Timestamp time.Time      `json:"timestamp"`
	TFState   config.TFState `json:"tfstate"`
}

// NewPayload creates a new Payload with the current timestamp.
func NewPayload(diff *tfstate.StateDiff, tfs config.TFState) *Payload {
	return &Payload{
		Diffs: diff,
		Metadata: Metadata{
			Timestamp: time.Now().UTC(),
			TFState:   tfs,
		},
	}
}
