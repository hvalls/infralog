package backend

// Backend defines the interface for reading Terraform state files from various sources.
type Backend interface {
	// GetState retrieves the Terraform state file contents.
	GetState() ([]byte, error)

	// Name returns the backend type name for logging purposes.
	Name() string
}
