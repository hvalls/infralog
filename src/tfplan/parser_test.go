package tfplan

import (
	"testing"
)

func TestParsePlan(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid plan with resources",
			json: `{
				"format_version": "1.2",
				"terraform_version": "1.5.0",
				"resource_changes": [{
					"address": "aws_instance.web",
					"mode": "managed",
					"type": "aws_instance",
					"name": "web",
					"change": {
						"actions": ["create"],
						"before": null,
						"after": {"ami": "ami-123"}
					}
				}]
			}`,
			wantErr: false,
		},
		{
			name: "empty plan",
			json: `{
				"format_version": "1.2",
				"resource_changes": []
			}`,
			wantErr: false,
		},
		{
			name: "missing format_version",
			json: `{
				"terraform_version": "1.5.0",
				"resource_changes": []
			}`,
			wantErr: true,
			errMsg:  "missing required field 'format_version'",
		},
		{
			name:    "invalid json",
			json:    `{"format_version": invalid}`,
			wantErr: true,
			errMsg:  "failed to parse plan JSON",
		},
		{
			name:    "empty string",
			json:    ``,
			wantErr: true,
			errMsg:  "failed to parse plan JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := ParsePlan([]byte(tt.json))
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParsePlan() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error()[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("ParsePlan() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("ParsePlan() unexpected error = %v", err)
				return
			}

			if plan == nil {
				t.Errorf("ParsePlan() returned nil plan")
				return
			}

			if plan.FormatVersion == "" {
				t.Errorf("ParsePlan() plan missing format_version")
			}
		})
	}
}

func TestParsePlanFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid plan file",
			filename: "testdata/plan_valid.json",
			wantErr:  false,
		},
		{
			name:     "empty plan file",
			filename: "testdata/plan_empty.json",
			wantErr:  false,
		},
		{
			name:     "invalid json file",
			filename: "testdata/plan_invalid.json",
			wantErr:  true,
			errMsg:   "failed to parse plan JSON",
		},
		{
			name:     "missing format_version",
			filename: "testdata/plan_no_version.json",
			wantErr:  true,
			errMsg:   "missing required field 'format_version'",
		},
		{
			name:     "non-existent file",
			filename: "testdata/does_not_exist.json",
			wantErr:  true,
			errMsg:   "failed to read plan file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := ParsePlanFile(tt.filename)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParsePlanFile() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error()[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("ParsePlanFile() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("ParsePlanFile() unexpected error = %v", err)
				return
			}

			if plan == nil {
				t.Errorf("ParsePlanFile() returned nil plan")
			}
		})
	}
}

func TestParsePlanFileValidStructure(t *testing.T) {
	plan, err := ParsePlanFile("testdata/plan_valid.json")
	if err != nil {
		t.Fatalf("ParsePlanFile() unexpected error = %v", err)
	}

	if plan.FormatVersion != "1.2" {
		t.Errorf("FormatVersion = %v, want 1.2", plan.FormatVersion)
	}

	if plan.TerraformVersion != "1.5.0" {
		t.Errorf("TerraformVersion = %v, want 1.5.0", plan.TerraformVersion)
	}

	if len(plan.ResourceChanges) != 1 {
		t.Errorf("ResourceChanges length = %v, want 1", len(plan.ResourceChanges))
	}

	if len(plan.ResourceChanges) > 0 {
		rc := plan.ResourceChanges[0]
		if rc.Type != "aws_instance" {
			t.Errorf("ResourceChange Type = %v, want aws_instance", rc.Type)
		}
		if rc.Name != "web" {
			t.Errorf("ResourceChange Name = %v, want web", rc.Name)
		}
		if len(rc.Change.Actions) != 1 || rc.Change.Actions[0] != "create" {
			t.Errorf("ResourceChange Actions = %v, want [create]", rc.Change.Actions)
		}
	}

	if len(plan.OutputChanges) != 1 {
		t.Errorf("OutputChanges length = %v, want 1", len(plan.OutputChanges))
	}
}

// TestParsePlanEmptyFile ensures empty resource/output changes are handled
func TestParsePlanEmptyFile(t *testing.T) {
	plan, err := ParsePlanFile("testdata/plan_empty.json")
	if err != nil {
		t.Fatalf("ParsePlanFile() unexpected error = %v", err)
	}

	if len(plan.ResourceChanges) != 0 {
		t.Errorf("Empty plan ResourceChanges length = %v, want 0", len(plan.ResourceChanges))
	}

	if len(plan.OutputChanges) != 0 {
		t.Errorf("Empty plan OutputChanges length = %v, want 0", len(plan.OutputChanges))
	}
}
