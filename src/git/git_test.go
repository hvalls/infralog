package git

import (
	"os"
	"testing"
)

func TestExtract_InGitRepo(t *testing.T) {
	// This test runs in the actual infralog git repo
	metadata := Extract()

	// In a git repository, we should get metadata
	// Note: Some fields might be empty depending on git configuration
	if metadata == nil {
		t.Fatal("Expected metadata but got nil (are we in a git repo?)")
	}

	// Commit SHA and Branch should always be available in a git repo
	if metadata.CommitSHA == "" {
		t.Error("Expected commit SHA to be non-empty in git repo")
	}

	if metadata.Branch == "" {
		t.Error("Expected branch name to be non-empty in git repo")
	}

	// Committer and RepoURL might be empty depending on git config
	// so we don't fail the test if they're missing
	t.Logf("Extracted metadata: Committer=%q, CommitSHA=%q, Branch=%q, RepoURL=%q",
		metadata.Committer, metadata.CommitSHA, metadata.Branch, metadata.RepoURL)
}

func TestExtract_NotInGitRepo(t *testing.T) {
	// Create a temp directory and run Extract from there
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	metadata := Extract()

	// Should return nil when not in a git repo
	if metadata != nil {
		t.Errorf("Expected nil metadata outside git repo, got: %+v", metadata)
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		metadata *Metadata
		want     bool
	}{
		{
			name:     "all empty",
			metadata: &Metadata{},
			want:     true,
		},
		{
			name: "committer only",
			metadata: &Metadata{
				Committer: "John Doe",
			},
			want: false,
		},
		{
			name: "commit sha only",
			metadata: &Metadata{
				CommitSHA: "abc123",
			},
			want: false,
		},
		{
			name: "branch only",
			metadata: &Metadata{
				Branch: "main",
			},
			want: false,
		},
		{
			name: "repo url only",
			metadata: &Metadata{
				RepoURL: "https://github.com/user/repo",
			},
			want: false,
		},
		{
			name: "all fields populated",
			metadata: &Metadata{
				Committer: "John Doe",
				CommitSHA: "abc123",
				Branch:    "main",
				RepoURL:   "https://github.com/user/repo",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.metadata.isEmpty(); got != tt.want {
				t.Errorf("isEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsGitAvailable(t *testing.T) {
	// This test verifies that isGitAvailable works
	// We expect git to be available in CI/development environments
	available := isGitAvailable()

	// We can't assert true/false definitively, but we can log it
	t.Logf("Git available: %v", available)

	// Just verify it doesn't panic
	if !available {
		t.Log("Warning: git is not available in this environment")
	}
}

func TestRunGitCommand(t *testing.T) {
	// Test that runGitCommand handles errors silently
	// Run a git command that should fail
	result := runGitCommand("invalid-command-that-does-not-exist")

	// Should return empty string on error (not panic)
	if result != "" {
		t.Errorf("Expected empty string on error, got: %q", result)
	}
}
