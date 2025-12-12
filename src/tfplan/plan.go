package tfplan

// Plan represents the structure of a Terraform plan JSON output.
// This matches the format produced by `terraform show -json plan.tfplan`.
type Plan struct {
	FormatVersion    string                  `json:"format_version"`
	TerraformVersion string                  `json:"terraform_version,omitempty"`
	ResourceChanges  []ResourceChange        `json:"resource_changes,omitempty"`
	OutputChanges    map[string]OutputChange `json:"output_changes,omitempty"`
	Configuration    map[string]interface{}  `json:"configuration,omitempty"`
	PlanningOptions  map[string]interface{}  `json:"planning_options,omitempty"`
}

// ResourceChange represents a single resource change in the plan.
type ResourceChange struct {
	Address       string `json:"address"`
	Mode          string `json:"mode"` // "managed" or "data"
	Type          string `json:"type"` // e.g., "aws_instance"
	Name          string `json:"name"`
	ProviderName  string `json:"provider_name,omitempty"`
	ModuleAddress string `json:"module_address,omitempty"`
	Change        Change `json:"change"`
	ActionReason  string `json:"action_reason,omitempty"`
}

// OutputChange represents a change to a Terraform output value.
type OutputChange struct {
	Change Change `json:"change"`
}

// Change describes the planned change for a resource or output.
type Change struct {
	Actions         []string               `json:"actions"`
	Before          map[string]interface{} `json:"before"`
	After           map[string]interface{} `json:"after"`
	AfterUnknown    map[string]interface{} `json:"after_unknown,omitempty"`
	BeforeSensitive interface{}            `json:"before_sensitive,omitempty"`
	AfterSensitive  interface{}            `json:"after_sensitive,omitempty"`
}
