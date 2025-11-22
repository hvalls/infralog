package persistence

import (
	"infralog/tfstate"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNewFileStore(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path",
			path:    filepath.Join(t.TempDir(), "state.json"),
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "nested path creates directories",
			path:    filepath.Join(t.TempDir(), "nested", "dir", "state.json"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewFileStore(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFileStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && store == nil {
				t.Error("NewFileStore() returned nil store without error")
			}
		})
	}
}

func TestFileStore_LoadNonExistent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	store, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	state, err := store.Load()
	if err != nil {
		t.Errorf("Load() error = %v, want nil for non-existent file", err)
	}
	if state != nil {
		t.Errorf("Load() = %v, want nil for non-existent file", state)
	}
}

func TestFileStore_SaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	want := &tfstate.State{
		Version:          4,
		TerraformVersion: "1.5.0",
		Serial:           42,
		Lineage:          "test-lineage",
		Resources: []tfstate.Resource{
			{
				Type: "aws_instance",
				Name: "web",
				Mode: "managed",
				Instances: []tfstate.ResourceInstance{
					{
						Attributes: map[string]any{
							"id":            "i-123456",
							"instance_type": "t2.micro",
						},
					},
				},
			},
		},
		Outputs: map[string]tfstate.Output{
			"instance_ip": {
				Value: "10.0.0.1",
			},
		},
	}

	if err := store.Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load() = %v, want %v", got, want)
	}
}

func TestFileStore_SaveOverwrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	first := &tfstate.State{Serial: 1}
	second := &tfstate.State{Serial: 2}

	if err := store.Save(first); err != nil {
		t.Fatalf("Save(first) error = %v", err)
	}

	if err := store.Save(second); err != nil {
		t.Fatalf("Save(second) error = %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got.Serial != 2 {
		t.Errorf("Load().Serial = %d, want 2", got.Serial)
	}
}

func TestFileStore_LoadCorruptedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	if err := os.WriteFile(path, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = store.Load()
	if err == nil {
		t.Error("Load() error = nil, want error for corrupted file")
	}
}

func TestFileStore_AtomicWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	state := &tfstate.State{Serial: 1}
	if err := store.Save(state); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify no temp file is left behind
	tempPath := path + ".tmp"
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("temporary file was not cleaned up")
	}

	// Verify the actual file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("state file was not created")
	}
}
