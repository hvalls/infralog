package tfstate

const (
	DiffStatusAdded     = "added"
	DiffStatusRemoved   = "removed"
	DiffStatusChanged   = "changed"
	DiffStatusUnchanged = "unchanged"
)

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
	Before any `json:"before,omitempty"`
	After  any `json:"after,omitempty"`
}

type StateDiff struct {
	ResourceDiffs []ResourceDiff `json:"resource_diffs"`
	OutputDiffs   []OutputDiff   `json:"output_diffs,omitempty"`
}

func (d *StateDiff) HasChanges() bool {
	return len(d.ResourceDiffs) > 0 || len(d.OutputDiffs) > 0
}
