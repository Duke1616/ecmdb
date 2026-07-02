package plugin

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Duke1616/ecmdb/internal/domain"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
	"github.com/samber/lo"
)

func filterGroups(filters []pluginx.Filter) []domain.FilterGroup {
	if len(filters) == 0 {
		return nil
	}

	conditions := lo.FilterMap(filters, func(filter pluginx.Filter, _ int) (domain.FilterCondition, bool) {
		if filter.Field == "" {
			return domain.FilterCondition{}, false
		}
		return domain.FilterCondition{
			FieldUID: filter.Field,
			Operator: domain.Operator(filter.Operator),
			Value:    filter.Value,
		}, true
	})
	if len(conditions) == 0 {
		return nil
	}
	return []domain.FilterGroup{{Filters: conditions}}
}

func filterResources(resources []domain.Resource, spec pluginx.ResourceSpec) []domain.Resource {
	if len(resources) == 0 {
		return resources
	}

	return lo.Filter(resources, func(resource domain.Resource, _ int) bool {
		if spec.ModelUID != "" && resource.ModelUID != spec.ModelUID {
			return false
		}
		return resourceMatchesFilters(resource, spec.Filters)
	})
}

func resourceMatchesFilters(resource domain.Resource, filters []pluginx.Filter) bool {
	return lo.EveryBy(filters, func(filter pluginx.Filter) bool {
		if filter.Field == "" {
			return true
		}
		return matchFilterValue(resource.Data[filter.Field], filter.Operator, filter.Value)
	})
}

func matchFilterValue(actual any, operator string, expected any) bool {
	switch strings.ToLower(operator) {
	case "", string(domain.OperatorEq):
		return fmt.Sprint(actual) == fmt.Sprint(expected)
	case string(domain.OperatorNe):
		return fmt.Sprint(actual) != fmt.Sprint(expected)
	case string(domain.OperatorContains):
		return strings.Contains(strings.ToLower(fmt.Sprint(actual)), strings.ToLower(fmt.Sprint(expected)))
	case string(domain.OperatorGt):
		return toFloat(actual) > toFloat(expected)
	case string(domain.OperatorLt):
		return toFloat(actual) < toFloat(expected)
	case "gte":
		return toFloat(actual) >= toFloat(expected)
	case "lte":
		return toFloat(actual) <= toFloat(expected)
	case "in":
		return containsValue(expected, actual)
	case "nin":
		return !containsValue(expected, actual)
	default:
		return fmt.Sprint(actual) == fmt.Sprint(expected)
	}
}

func containsValue(list any, value any) bool {
	rv := reflect.ValueOf(list)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return fmt.Sprint(list) == fmt.Sprint(value)
	}
	for i := 0; i < rv.Len(); i++ {
		if fmt.Sprint(rv.Index(i).Interface()) == fmt.Sprint(value) {
			return true
		}
	}
	return false
}

func toFloat(value any) float64 {
	switch v := value.(type) {
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		return 0
	}
}
