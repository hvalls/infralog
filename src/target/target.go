package target

import (
	"infralog/git"
	"infralog/tfplan"
	"time"
)

// Target defines the interface for notification targets.
type Target interface {
	Write(*Payload) error
}

// Payload contains the change data sent to targets.
type Payload struct {
	Plan     *tfplan.Plan     `json:"plan"`
	Datetime time.Time        `json:"datetime"`
	Metadata *PayloadMetadata `json:"metadata,omitempty"`
}

// PayloadMetadata contains additional context about the infrastructure change.
type PayloadMetadata struct {
	Git *git.Metadata `json:"git,omitempty"`
}

// NewPayload creates a new Payload with the current timestamp and metadata.
func NewPayload(plan *tfplan.Plan) *Payload {
	return &Payload{
		Plan:     plan,
		Datetime: time.Now().UTC(),
		Metadata: extractMetadata(),
	}
}

// extractMetadata attempts to extract metadata from the environment.
// Returns nil if no metadata is available.
func extractMetadata() *PayloadMetadata {
	gitMeta := git.Extract()
	if gitMeta == nil {
		return nil
	}

	return &PayloadMetadata{
		Git: gitMeta,
	}
}
