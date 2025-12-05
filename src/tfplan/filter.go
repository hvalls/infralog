package tfplan

import (
	"infralog/config"
	"sort"
)

// ApplyFilter creates a new Plan with filters applied to resource and output changes.
// It removes resources that don't match the filter and actions that should be ignored.
func ApplyFilter(plan *Plan, filter config.Filter) *Plan {
	filtered := &Plan{
		FormatVersion:    plan.FormatVersion,
		TerraformVersion: plan.TerraformVersion,
		Configuration:    plan.Configuration,
		PlanningOptions:  plan.PlanningOptions,
		ResourceChanges:  []ResourceChange{},
		OutputChanges:    make(map[string]OutputChange),
	}

	// Filter resource changes
	for _, rc := range plan.ResourceChanges {
		if shouldIncludeResource(rc, filter) {
			filtered.ResourceChanges = append(filtered.ResourceChanges, rc)
		}
	}

	// Filter output changes
	for name, oc := range plan.OutputChanges {
		if shouldIncludeOutput(name, oc, filter) {
			filtered.OutputChanges[name] = oc
		}
	}

	return filtered
}

// HasChanges returns true if the plan contains any resource or output changes.
func (p *Plan) HasChanges() bool {
	return len(p.ResourceChanges) > 0 || len(p.OutputChanges) > 0
}

// shouldIncludeResource determines if a resource change should be included.
func shouldIncludeResource(rc ResourceChange, filter config.Filter) bool {
	// Apply resource type filter
	if !filter.MatchesResourceType(rc.Type) {
		return false
	}

	// Check if actions should be included
	return shouldIncludeActions(rc.Change.Actions)
}

// shouldIncludeOutput determines if an output change should be included.
func shouldIncludeOutput(name string, oc OutputChange, filter config.Filter) bool {
	// Apply output name filter
	if !filter.MatchesOutput(name) {
		return false
	}

	// Check if actions should be included
	return shouldIncludeActions(oc.Change.Actions)
}

// shouldIncludeActions returns true if the actions represent a meaningful change.
// Filters out "no-op" and "read" actions.
func shouldIncludeActions(actions []string) bool {
	if len(actions) == 0 {
		return false
	}

	// Sort actions to normalize ordering
	sortedActions := make([]string, len(actions))
	copy(sortedActions, actions)
	sort.Strings(sortedActions)

	// Filter out no-ops and reads
	for _, action := range sortedActions {
		if action == "no-op" || action == "read" {
			return false
		}
	}

	// If we have any other actions (create, delete, update), include it
	return true
}
