package services

import (
	"github.com/mudler/LocalAGI/core/state"
	"github.com/mudler/LocalAGI/core/types"
	"github.com/mudler/LocalAGI/pkg/config"
	"github.com/mudler/LocalAGI/services/filters"
)

// Filters loads all filters from agent config.
func Filters(a *state.AgentConfig) types.JobFilters {
	var result []types.JobFilter
	for _, f := range a.Filters {
		switch f.Type {
		case filters.FilterRegex:
			filter, err := filters.NewRegexFilter(f.Config)
			if err != nil {
				// log error if needed
				continue
			}
			result = append(result, filter)
			// Add other filter types here
		}
	}
	return result
}

// FiltersConfigMeta returns all filter config metas for UI.
func FiltersConfigMeta() []config.FieldGroup {
	return []config.FieldGroup{
		filters.RegexFilterConfigMeta(),
		// Add other filter config metas here
	}
}
