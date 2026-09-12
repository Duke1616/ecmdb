package codec

import (
	"reflect"
	"strings"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
)

type pluginTag struct {
	name         string
	label        string // 中文展示名，优先作为 Attribute.Name（不影响 UID 映射）
	model        string
	relationType string
	direction    string
	cardinality  string
	required     bool
	field        string
	defaultValue string
	skip         bool
}

// displayName 返回最终展示用的名称：label 优先，其次 name
func (t pluginTag) displayName() string {
	if t.label != "" {
		return t.label
	}
	return t.name
}

func parsePluginTag(field reflect.StructField) pluginTag {
	raw := field.Tag.Get("plugin")
	if raw == "-" {
		return pluginTag{skip: true}
	}
	if raw == "" {
		return pluginTag{name: fallbackFieldName(field)}
	}

	parts := strings.Split(raw, ",")
	tag := pluginTag{}

	first := strings.TrimSpace(parts[0])
	if first != "" && !strings.Contains(first, "=") {
		tag.name = first
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
			switch k {
			case "name":
				tag.name = v
			case "label", "title":
				// label/title 均作为中文展示名
				tag.label = v
			case "model":
				tag.model = v
			case "type", "relation_type":
				tag.relationType = v
			case "direction":
				tag.direction = v
			case "in":
				tag.relationType = v
				tag.direction = types.DirectionToSource
			case "out":
				tag.relationType = v
				tag.direction = types.DirectionToTarget
			case "cardinality":
				tag.cardinality = v
			case "field":
				tag.field = v
			case "default":
				tag.defaultValue = v
			}
			continue
		}
		if part == "required" {
			tag.required = true
		}
	}

	if tag.name == "" {
		tag.name = fallbackFieldName(field)
	}
	return tag
}

func fallbackFieldName(field reflect.StructField) string {
	if raw := field.Tag.Get("json"); raw != "" {
		name := strings.Split(raw, ",")[0]
		if name != "" && name != "-" {
			return name
		}
	}
	return strings.ToLower(field.Name[:1]) + field.Name[1:]
}
