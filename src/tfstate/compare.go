package tfstate

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func Compare(oldState, newState *State) (*StateDiff, error) {
	stateDiff := &StateDiff{}

	if oldState == nil || newState == nil {
		return nil, fmt.Errorf("oldState and newState cannot be nil")
	}

	stateDiff.MetadataChanged = (oldState.Version != newState.Version) ||
		(oldState.TerraformVersion != newState.TerraformVersion) ||
		(oldState.Lineage != newState.Lineage)

	oldResourceMap := make(map[string]Resource)
	newResourceMap := make(map[string]Resource)

	for _, resource := range oldState.Resources {
		oldResourceMap[ResourceID(resource)] = resource
	}

	for _, resource := range newState.Resources {
		newResourceMap[ResourceID(resource)] = resource
	}

	allResourceIDs := make(map[string]bool)
	for id := range oldResourceMap {
		allResourceIDs[id] = true
	}
	for id := range newResourceMap {
		allResourceIDs[id] = true
	}

	var sortedResourceIDs []string
	for id := range allResourceIDs {
		sortedResourceIDs = append(sortedResourceIDs, id)
	}
	sort.Strings(sortedResourceIDs)

	for _, id := range sortedResourceIDs {
		oldResource, oldExists := oldResourceMap[id]
		newResource, newExists := newResourceMap[id]

		var resourceDiff ResourceDiff
		parts := splitResourceID(id)
		resourceDiff.ResourceType = parts.resourceType
		resourceDiff.ResourceName = parts.resourceName

		if !oldExists {
			resourceDiff.Status = DiffStatusAdded
			stateDiff.ResourceDiffs = append(stateDiff.ResourceDiffs, resourceDiff)
			continue
		}

		if !newExists {
			resourceDiff.Status = DiffStatusRemoved
			stateDiff.ResourceDiffs = append(stateDiff.ResourceDiffs, resourceDiff)
			continue
		}

		resourceDiff.Status = DiffStatusUnchanged
		resourceDiff.AttributeDiffs = make(map[string]ValueDiff)

		attrDiffs := compareInstances(oldResource.Instances, newResource.Instances)
		if len(attrDiffs) > 0 {
			resourceDiff.Status = DiffStatusChanged
			resourceDiff.AttributeDiffs = attrDiffs
			stateDiff.ResourceDiffs = append(stateDiff.ResourceDiffs, resourceDiff)
		} else if resourceDiff.Status != DiffStatusUnchanged {
			stateDiff.ResourceDiffs = append(stateDiff.ResourceDiffs, resourceDiff)
		}
	}

	if oldState.Outputs != nil || newState.Outputs != nil {
		outputDiffs := compareOutputs(oldState.Outputs, newState.Outputs)
		if len(outputDiffs) > 0 {
			stateDiff.OutputDiffs = outputDiffs
		}
	}

	return stateDiff, nil
}

// splitResourceID splits a resource ID into its component parts
func splitResourceID(id string) ResourceIDParts {
	var parts ResourceIDParts

	segments := strings.Split(id, ".")

	if len(segments) < 2 {
		// Invalid format, return empty parts
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

// compareInstances compares the instances of a resource
func compareInstances(oldInstances, newInstances []ResourceInstance) map[string]ValueDiff {
	attrDiffs := make(map[string]ValueDiff)

	// Simple case: compare first instance (for non-count resources)
	if len(oldInstances) > 0 && len(newInstances) > 0 {
		oldAttrs := oldInstances[0].Attributes
		newAttrs := newInstances[0].Attributes

		// Get all attribute keys
		allKeys := make(map[string]bool)
		for k := range oldAttrs {
			allKeys[k] = true
		}
		for k := range newAttrs {
			allKeys[k] = true
		}

		// Compare each attribute
		for key := range allKeys {
			oldVal, oldExists := oldAttrs[key]
			newVal, newExists := newAttrs[key]

			if !oldExists {
				attrDiffs[key] = ValueDiff{NewValue: newVal}
			} else if !newExists {
				attrDiffs[key] = ValueDiff{OldValue: oldVal}
			} else if !reflect.DeepEqual(oldVal, newVal) {
				attrDiffs[key] = ValueDiff{
					OldValue: oldVal,
					NewValue: newVal,
				}
			}
		}
	}

	return attrDiffs
}

func compareOutputs(oldOutputs, newOutputs map[string]Output) []OutputDiff {
	var outputDiffs []OutputDiff

	// Get all output names
	allNames := make(map[string]bool)
	for name := range oldOutputs {
		allNames[name] = true
	}
	for name := range newOutputs {
		allNames[name] = true
	}

	var sortedNames []string
	for name := range allNames {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	for _, name := range sortedNames {
		oldOutput, oldExists := oldOutputs[name]
		newOutput, newExists := newOutputs[name]

		var outputDiff OutputDiff
		outputDiff.OutputName = name

		if !oldExists {
			outputDiff.Status = DiffStatusAdded
			outputDiffs = append(outputDiffs, outputDiff)
			continue
		}

		if !newExists {
			outputDiff.Status = DiffStatusRemoved
			outputDiffs = append(outputDiffs, outputDiff)
			continue
		}

		if !reflect.DeepEqual(oldOutput.Value, newOutput.Value) {
			outputDiff.Status = DiffStatusChanged
			outputDiff.ValueDiff = ValueDiff{
				OldValue: oldOutput.Value,
				NewValue: newOutput.Value,
			}
			outputDiffs = append(outputDiffs, outputDiff)
		}
	}

	return outputDiffs
}
