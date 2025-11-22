package local

import (
	"infralog/config"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	cfg := config.LocalConfig{Path: "/path/to/state.tfstate"}
	backend := New(cfg)

	if backend.path != cfg.Path {
		t.Errorf("Expected path %q, got %q", cfg.Path, backend.path)
	}
}

func TestName(t *testing.T) {
	backend := &LocalBackend{}
	if backend.Name() != "local" {
		t.Errorf("Expected name 'local', got %q", backend.Name())
	}
}

func TestGetState_Success(t *testing.T) {
	// Create a temporary file with test content
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "terraform.tfstate")
	expectedContent := `{"version": 4, "resources": []}`

	if err := os.WriteFile(statePath, []byte(expectedContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	backend := New(config.LocalConfig{Path: statePath})
	data, err := backend.GetState()

	if err != nil {
		t.Fatalf("GetState() returned error: %v", err)
	}

	if string(data) != expectedContent {
		t.Errorf("GetState() = %q, expected %q", string(data), expectedContent)
	}
}

func TestGetState_FileNotFound(t *testing.T) {
	backend := New(config.LocalConfig{Path: "/nonexistent/path/terraform.tfstate"})
	_, err := backend.GetState()

	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestGetState_PermissionDenied(t *testing.T) {
	// Create a temporary file with no read permissions
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "terraform.tfstate")

	if err := os.WriteFile(statePath, []byte("content"), 0000); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	backend := New(config.LocalConfig{Path: statePath})
	_, err := backend.GetState()

	if err == nil {
		t.Error("Expected error for unreadable file, got nil")
	}
}
