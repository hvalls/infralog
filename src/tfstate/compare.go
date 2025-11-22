package tfstate

import (
	"fmt"
	"infralog/config"
	"reflect"
	"sort"
)

func Compare(oldState, newState *State, filter config.Filter) (*StateDiff, error) {
	stateDiff := &StateDiff{}

	if oldState == nil || newState == nil {
		return nil, fmt.Errorf("oldState and newState cannot be nil")
	}

	oldResourceMap := mapResources(oldState.Resources, filter)
	newResourceMap := mapResources(newState.Resources, filter)

	allResourceIDs := make(map[ResourceID]bool)
	for id := range oldResourceMap {
		allResourceIDs[id] = true
	}
	for id := range newResourceMap {
		allResourceIDs[id] = true
	}

	var sortedResourceIDs []ResourceID
	for id := range allResourceIDs {
		sortedResourceIDs = append(sortedResourceIDs, id)
	}
	sort.Slice(sortedResourceIDs, func(i, j int) bool {
		return string(sortedResourceIDs[i]) < string(sortedResourceIDs[j])
	})

	for _, id := range sortedResourceIDs {
		oldResource, oldExists := oldResourceMap[id]
		newResource, newExists := newResourceMap[id]

		var resourceDiff ResourceDiff
		parts := id.Split()
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

	filteredOldOutputs := make(map[string]Output)
	for name, output := range oldState.Outputs {
		if filter.MatchesOutput(name) {
			filteredOldOutputs[name] = output
		}
	}

	filteredNewOutputs := make(map[string]Output)
	for name, output := range newState.Outputs {
		if filter.MatchesOutput(name) {
			filteredNewOutputs[name] = output
		}
	}

	outputDiffs := compareOutputs(filteredOldOutputs, filteredNewOutputs)
	if len(outputDiffs) > 0 {
		stateDiff.OutputDiffs = outputDiffs
	}

	return stateDiff, nil
}

func mapResources(resources []Resource, filter config.Filter) map[ResourceID]Resource {
	resourceMap := make(map[ResourceID]Resource)
	for _, resource := range resources {
		id := resource.GetID()
		parts := id.Split()
		if filter.MatchesResourceType(parts.resourceType) {
			resourceMap[id] = resource
		}
	}
	return resourceMap
}

func compareInstances(oldInstances, newInstances []ResourceInstance) map[string]ValueDiff {
	attrDiffs := make(map[string]ValueDiff)

	if len(oldInstances) > 0 && len(newInstances) > 0 {
		oldAttrs := oldInstances[0].Attributes
		newAttrs := newInstances[0].Attributes

		allKeys := make(map[string]bool)
		for k := range oldAttrs {
			allKeys[k] = true
		}
		for k := range newAttrs {
			allKeys[k] = true
		}

		for key := range allKeys {
			oldVal, oldExists := oldAttrs[key]
			newVal, newExists := newAttrs[key]

			if !oldExists {
				attrDiffs[key] = ValueDiff{After: newVal}
			} else if !newExists {
				attrDiffs[key] = ValueDiff{Before: oldVal}
			} else if !reflect.DeepEqual(oldVal, newVal) {
				attrDiffs[key] = ValueDiff{
					Before: oldVal,
					After:  newVal,
				}
			}
		}
	}

	return attrDiffs
}

func compareOutputs(oldOutputs, newOutputs map[string]Output) []OutputDiff {
	var outputDiffs []OutputDiff

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
				Before: oldOutput.Value,
				After:  newOutput.Value,
			}
			outputDiffs = append(outputDiffs, outputDiff)
		}
	}

	return outputDiffs
}
