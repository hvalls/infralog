// Package persistence provides state storage for Infralog.
//
// It allows persisting the last-seen Terraform state to disk so that
// state changes are not missed across restarts.
package persistence

import "infralog/tfstate"

// Store defines the interface for persisting Terraform state.
// Implementations must be safe for concurrent use.
type Store interface {
	// Load retrieves the last persisted state.
	// Returns nil, nil if no state has been persisted yet.
	Load() (*tfstate.State, error)

	// Save persists the given state.
	Save(state *tfstate.State) error
}
