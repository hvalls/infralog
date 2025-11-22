package local

import (
	"fmt"
	"infralog/config"
	"os"
)

type LocalBackend struct {
	path string
}

func New(cfg config.LocalConfig) *LocalBackend {
	return &LocalBackend{
		path: cfg.Path,
	}
}

func (b *LocalBackend) GetState() ([]byte, error) {
	data, err := os.ReadFile(b.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}
	return data, nil
}

func (b *LocalBackend) Name() string {
	return "local"
}
