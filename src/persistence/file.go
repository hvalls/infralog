package persistence

import (
	"encoding/json"
	"errors"
	"fmt"
	"infralog/tfstate"
	"os"
	"path/filepath"
	"sync"
)

// FileStore persists state to a local JSON file.
// It uses atomic writes to prevent corruption.
type FileStore struct {
	path string
	mu   sync.Mutex
}

// NewFileStore creates a new FileStore that persists state to the given path.
// The parent directory will be created if it doesn't exist.
func NewFileStore(path string) (*FileStore, error) {
	if path == "" {
		return nil, errors.New("persistence path cannot be empty")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create persistence directory: %w", err)
	}

	return &FileStore{path: path}, nil
}

// Load reads the persisted state from disk.
// Returns nil, nil if the file doesn't exist yet.
func (f *FileStore) Load() (*tfstate.State, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := os.ReadFile(f.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read persisted state: %w", err)
	}

	var state tfstate.State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse persisted state: %w", err)
	}

	return &state, nil
}

// Save writes the state to disk atomically.
// It first writes to a temporary file, then renames it to the target path.
func (f *FileStore) Save(state *tfstate.State) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to a temporary file first for atomic operation
	tempPath := f.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary state file: %w", err)
	}

	// Rename is atomic on POSIX systems
	if err := os.Rename(tempPath, f.path); err != nil {
		os.Remove(tempPath) // Clean up on failure
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	return nil
}
