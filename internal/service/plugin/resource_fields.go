package plugin

import (
	"github.com/Duke1616/ecmdb/internal/domain"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
	"github.com/samber/lo"
)

func specFields(spec pluginx.ResourceSpec) []string {
	fields := lo.Values(spec.Fields)
	fields = append(fields, lo.Map(spec.Filters, func(filter pluginx.Filter, _ int) string {
		return filter.Field
	})...)
	return lo.Uniq(lo.Filter(fields, func(field string, _ int) bool {
		return field != ""
	}))
}

func resolveFields(resource domain.Resource, spec pluginx.ResourceSpec) map[string]any {
	return lo.MapValues(spec.Fields, func(field string, _ string) any {
		return resource.Data[field]
	})
}

func requiredFieldsSatisfied(requirements []string, fields map[string]any) bool {
	return lo.EveryBy(requirements, func(fieldName string) bool {
		return hasValue(fields[fieldName])
	})
}

func hasValue(value any) bool {
	switch v := value.(type) {
	case nil:
		return false
	case string:
		return v != ""
	default:
		return true
	}
}
