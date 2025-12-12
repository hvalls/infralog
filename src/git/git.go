package git

import (
	"os/exec"
	"strings"
)

// Metadata contains git repository information automatically extracted
// from the local git repository.
type Metadata struct {
	Committer string `json:"committer,omitempty"`
	CommitSHA string `json:"commit_sha,omitempty"`
	Branch    string `json:"branch,omitempty"`
	RepoURL   string `json:"repo_url,omitempty"`
}

// Extract attempts to extract git metadata from the current directory.
// Returns nil if git is not available or not in a git repository.
// All errors are silently ignored per requirements.
func Extract() *Metadata {
	// Check if git is available
	if !isGitAvailable() {
		return nil
	}

	// Check if we're in a git repository by trying to get the commit SHA
	// This is the most reliable indicator of being in a git repo
	commitSHA := runGitCommand("rev-parse", "HEAD")
	if commitSHA == "" {
		// Not in a git repository
		return nil
	}

	metadata := &Metadata{
		Committer: runGitCommand("config", "user.name"),
		CommitSHA: commitSHA,
		Branch:    runGitCommand("rev-parse", "--abbrev-ref", "HEAD"),
		RepoURL:   runGitCommand("config", "--get", "remote.origin.url"),
	}

	return metadata
}

// isGitAvailable checks if git is available in the system.
func isGitAvailable() bool {
	cmd := exec.Command("git", "--version")
	return cmd.Run() == nil
}

// runGitCommand executes a git command and returns the trimmed output.
// Returns empty string on error (silent failure).
func runGitCommand(args ...string) string {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// isEmpty checks if all metadata fields are empty.
func (m *Metadata) isEmpty() bool {
	return m.Committer == "" && m.CommitSHA == "" &&
		m.Branch == "" && m.RepoURL == ""
}
