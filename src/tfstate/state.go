package tfstate

import (
	"encoding/json"
	"fmt"
)

var LastState *State

type State struct {
	Version          int               `json:"version"`
	TerraformVersion string            `json:"terraform_version"`
	Serial           int               `json:"serial"`
	Lineage          string            `json:"lineage"`
	Resources        []Resource        `json:"resources"`
	Outputs          map[string]Output `json:"outputs,omitempty"`
}

func ParseState(data string) (*State, error) {
	var state State
	err := json.Unmarshal([]byte(data), &state)
	if err != nil {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}
	return &state, nil
}
