package tfplan

import (
	"infralog/config"
	"testing"
)

func TestApplyFilter(t *testing.T) {
	tests := []struct {
		name                string
		planFile            string
		filter              config.Filter
		wantResourceChanges int
		wantOutputChanges   int
		wantFirstResType    string
		wantFirstResAction  string
	}{
		{
			name:                "create action",
			planFile:            "testdata/plan_valid.json",
			filter:              config.Filter{}, // nil filter = match all
			wantResourceChanges: 1,
			wantOutputChanges:   1,
			wantFirstResType:    "aws_instance",
			wantFirstResAction:  "create",
		},
		{
			name:                "update action",
			planFile:            "testdata/plan_update.json",
			filter:              config.Filter{},
			wantResourceChanges: 1,
			wantOutputChanges:   0,
			wantFirstResType:    "aws_instance",
			wantFirstResAction:  "update",
		},
		{
			name:                "delete action",
			planFile:            "testdata/plan_delete.json",
			filter:              config.Filter{},
			wantResourceChanges: 1,
			wantOutputChanges:   0,
			wantFirstResType:    "aws_instance",
			wantFirstResAction:  "delete",
		},
		{
			name:                "replace action",
			planFile:            "testdata/plan_replace.json",
			filter:              config.Filter{},
			wantResourceChanges: 1,
			wantOutputChanges:   0,
			wantFirstResType:    "aws_instance",
		},
		{
			name:                "no-op action filtered out",
			planFile:            "testdata/plan_noop.json",
			filter:              config.Filter{},
			wantResourceChanges: 0,
			wantOutputChanges:   0,
		},
		{
			name:                "empty plan",
			planFile:            "testdata/plan_empty.json",
			filter:              config.Filter{},
			wantResourceChanges: 0,
			wantOutputChanges:   0,
		},
		{
			name:                "mixed actions",
			planFile:            "testdata/plan_mixed.json",
			filter:              config.Filter{},
			wantResourceChanges: 4,
			wantOutputChanges:   1,
			wantFirstResType:    "aws_instance",
			wantFirstResAction:  "create",
		},
		{
			name:     "filter by resource type",
			planFile: "testdata/plan_mixed.json",
			filter: config.Filter{
				ResourceTypes: []string{"aws_s3_bucket"},
			},
			wantResourceChanges: 1,
			wantOutputChanges:   1,
			wantFirstResType:    "aws_s3_bucket",
		},
		{
			name:     "empty filter blocks all resources",
			planFile: "testdata/plan_mixed.json",
			filter: config.Filter{
				ResourceTypes: []string{},
			},
			wantResourceChanges: 0,
			wantOutputChanges:   1,
		},
		{
			name:     "filter outputs",
			planFile: "testdata/plan_mixed.json",
			filter: config.Filter{
				ResourceTypes: []string{"aws_instance"},
				Outputs:       []string{"instance_ip"},
			},
			wantResourceChanges: 3,
			wantOutputChanges:   1,
		},
		{
			name:     "empty output filter blocks all outputs",
			planFile: "testdata/plan_mixed.json",
			filter: config.Filter{
				Outputs: []string{},
			},
			wantResourceChanges: 4,
			wantOutputChanges:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse plan file
			plan, err := ParsePlanFile(tt.planFile)
			if err != nil {
				t.Fatalf("ParsePlanFile() error = %v", err)
			}

			// Apply filter
			filtered := ApplyFilter(plan, tt.filter)

			// Check resource changes count
			if len(filtered.ResourceChanges) != tt.wantResourceChanges {
				t.Errorf("ResourceChanges count = %v, want %v", len(filtered.ResourceChanges), tt.wantResourceChanges)
			}

			// Check output changes count
			if len(filtered.OutputChanges) != tt.wantOutputChanges {
				t.Errorf("OutputChanges count = %v, want %v", len(filtered.OutputChanges), tt.wantOutputChanges)
			}

			// Check first resource change if expected
			if tt.wantResourceChanges > 0 && len(filtered.ResourceChanges) > 0 && tt.wantFirstResType != "" {
				first := filtered.ResourceChanges[0]
				if first.Type != tt.wantFirstResType {
					t.Errorf("First ResourceChange type = %v, want %v", first.Type, tt.wantFirstResType)
				}
				if tt.wantFirstResAction != "" && len(first.Change.Actions) > 0 && first.Change.Actions[0] != tt.wantFirstResAction {
					t.Errorf("First ResourceChange action = %v, want %v", first.Change.Actions[0], tt.wantFirstResAction)
				}
			}
		})
	}
}

func TestShouldIncludeActions(t *testing.T) {
	tests := []struct {
		name        string
		actions     []string
		wantInclude bool
	}{
		{
			name:        "create action",
			actions:     []string{"create"},
			wantInclude: true,
		},
		{
			name:        "delete action",
			actions:     []string{"delete"},
			wantInclude: true,
		},
		{
			name:        "update action",
			actions:     []string{"update"},
			wantInclude: true,
		},
		{
			name:        "no-op action filtered out",
			actions:     []string{"no-op"},
			wantInclude: false,
		},
		{
			name:        "read action filtered out",
			actions:     []string{"read"},
			wantInclude: false,
		},
		{
			name:        "replace (delete + create)",
			actions:     []string{"delete", "create"},
			wantInclude: true,
		},
		{
			name:        "replace (create + delete)",
			actions:     []string{"create", "delete"},
			wantInclude: true,
		},
		{
			name:        "empty actions",
			actions:     []string{},
			wantInclude: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			include := shouldIncludeActions(tt.actions)
			if include != tt.wantInclude {
				t.Errorf("shouldIncludeActions() = %v, want %v", include, tt.wantInclude)
			}
		})
	}
}

func TestPlanHasChanges(t *testing.T) {
	tests := []struct {
		name       string
		plan       *Plan
		wantResult bool
	}{
		{
			name: "has resource changes",
			plan: &Plan{
				ResourceChanges: []ResourceChange{
					{Type: "aws_instance", Name: "web", Change: Change{Actions: []string{"create"}}},
				},
			},
			wantResult: true,
		},
		{
			name: "has output changes",
			plan: &Plan{
				OutputChanges: map[string]OutputChange{
					"ip": {Change: Change{Actions: []string{"create"}}},
				},
			},
			wantResult: true,
		},
		{
			name: "has both",
			plan: &Plan{
				ResourceChanges: []ResourceChange{
					{Type: "aws_instance", Name: "web", Change: Change{Actions: []string{"create"}}},
				},
				OutputChanges: map[string]OutputChange{
					"ip": {Change: Change{Actions: []string{"create"}}},
				},
			},
			wantResult: true,
		},
		{
			name: "no changes",
			plan: &Plan{
				ResourceChanges: []ResourceChange{},
				OutputChanges:   map[string]OutputChange{},
			},
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.plan.HasChanges()
			if result != tt.wantResult {
				t.Errorf("HasChanges() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}
