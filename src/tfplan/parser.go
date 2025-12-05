package tfplan

import (
	"encoding/json"
	"fmt"
	"os"
)

// ParsePlanFile reads a Terraform plan JSON file and parses it into a Plan struct.
// Returns an error if the file cannot be read or if the JSON is invalid.
func ParsePlanFile(filename string) (*Plan, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read plan file: %w", err)
	}

	return ParsePlan(data)
}

// ParsePlan parses Terraform plan JSON data into a Plan struct.
// Returns an error if the JSON is invalid or missing required fields.
func ParsePlan(data []byte) (*Plan, error) {
	var plan Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
	}

	// Validate required fields
	if plan.FormatVersion == "" {
		return nil, fmt.Errorf("missing required field 'format_version' in plan JSON")
	}

	return &plan, nil
}
