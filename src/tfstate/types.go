package tfstate

import "fmt"

const (
	DiffStatusAdded     = "added"
	DiffStatusRemoved   = "removed"
	DiffStatusChanged   = "changed"
	DiffStatusUnchanged = "unchanged"
)

type State struct {
	Version          int               `json:"version"`
	TerraformVersion string            `json:"terraform_version"`
	Serial           int               `json:"serial"`
	Lineage          string            `json:"lineage"`
	Resources        []Resource        `json:"resources"`
	Outputs          map[string]Output `json:"outputs,omitempty"`
}

type Resource struct {
	Module    string             `json:"module,omitempty"`
	Mode      string             `json:"mode"`
	Type      string             `json:"type"`
	Name      string             `json:"name"`
	Provider  string             `json:"provider"`
	Instances []ResourceInstance `json:"instances"`
}

type ResourceIDParts struct {
	module       string
	resourceType string
	resourceName string
}

type ResourceInstance struct {
	IndexKey      any            `json:"index_key,omitempty"`
	SchemaVersion int            `json:"schema_version"`
	Attributes    map[string]any `json:"attributes"`
	Private       string         `json:"private,omitempty"`
}

type Output struct {
	Value     any  `json:"value"`
	Type      any  `json:"type,omitempty"`
	Sensitive bool `json:"sensitive,omitempty"`
}

type ResourceDiff struct {
	ResourceType   string               `json:"resource_type"`
	ResourceName   string               `json:"resource_name"`
	Status         ResourceDiffStatus   `json:"status"`
	AttributeDiffs map[string]ValueDiff `json:"attribute_diffs,omitempty"`
}

type OutputDiff struct {
	OutputName string             `json:"output_name"`
	Status     ResourceDiffStatus `json:"status"`
	ValueDiff  ValueDiff          `json:"value_diff"`
}

type ResourceDiffStatus string

type ValueDiff struct {
	OldValue any `json:"old_value,omitempty"`
	NewValue any `json:"new_value,omitempty"`
}

type StateDiff struct {
	MetadataChanged bool           `json:"metadata_changed"`
	ResourceDiffs   []ResourceDiff `json:"resource_diffs"`
	OutputDiffs     []OutputDiff   `json:"output_diffs,omitempty"`
}

func ResourceID(r Resource) string {
	modulePrefix := ""
	if r.Module != "" {
		modulePrefix = r.Module + "."
	}
	return fmt.Sprintf("%s%s.%s", modulePrefix, r.Type, r.Name)
}
