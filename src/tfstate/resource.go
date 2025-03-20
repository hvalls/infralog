package tfstate

import (
	"fmt"
	"strings"
)

type ResourceID string

type Resource struct {
	Module    string             `json:"module,omitempty"`
	Mode      string             `json:"mode"`
	Type      string             `json:"type"`
	Name      string             `json:"name"`
	Provider  string             `json:"provider"`
	Instances []ResourceInstance `json:"instances"`
}

type ResourceInstance struct {
	IndexKey      any            `json:"index_key,omitempty"`
	SchemaVersion int            `json:"schema_version"`
	Attributes    map[string]any `json:"attributes"`
	Private       string         `json:"private,omitempty"`
}

type ResourceIDParts struct {
	module       string
	resourceType string
	resourceName string
}

func (r Resource) GetID() ResourceID {
	modulePrefix := ""
	if r.Module != "" {
		modulePrefix = r.Module + "."
	}
	return ResourceID(fmt.Sprintf("%s%s.%s", modulePrefix, r.Type, r.Name))
}

func (id ResourceID) Split() ResourceIDParts {
	var parts ResourceIDParts

	segments := strings.Split(string(id), ".")

	if len(segments) < 2 {
		return parts
	}

	if segments[0] == "module" && len(segments) >= 4 {
		// Format: module.module_name.resource_type.resource_name
		moduleSegments := segments[:len(segments)-2]
		parts.module = strings.Join(moduleSegments, ".")
		parts.resourceType = segments[len(segments)-2]
		parts.resourceName = segments[len(segments)-1]
	} else {
		// Format: resource_type.resource_name
		parts.resourceType = segments[0]
		parts.resourceName = segments[1]
	}

	return parts
}
