package codec

import (
	"reflect"
	"strings"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
)

type pluginTag struct {
	key          string   // 字段程序标识键（如 host, gateways，取自首项或 Go 字段名）
	name         string   // 人类可读展示名称（统一由 name= 或 label= 显式声明，如 "主机地址"、"跳板机网关"）
	field        string   // CMDB 属性 UID 映射（如 field=ip，缺省回退到 key）
	model        string   // CMDB 子模型 UID（如 model=AuthGateway）
	group        string   // 所属 CMDB 模型分组名称（如 group=安全凭据）
	fieldType    string   // 显式声明的 CMDB 属性数据类型（如 list, string, number 等）
	options      []string // 枚举下拉选项列表（由 options= 或 enums= 声明，支持 | 分隔）
	relationType string   // 关联类型 UID（如 default, run）
	direction    string   // 关联方向（auto, source, target）
	cardinality  string   // 数量基数（one, many）
	required     bool     // 是否必填
	secure       *bool    // 是否敏感加密存储（nil 表示未显式声明，遵循关键字自动推导）
	defaultValue string   // 默认值
	skip         bool     // 是否忽略该字段
}

// CMDBUID 返回该属性在 CMDB 中的字段 UID（优先取 field，缺省回退取 key）
func (t pluginTag) CMDBUID() string {
	return lo.Ternary(t.field != "", t.field, t.key)
}

// DisplayName 返回展示名称（优先取 name，缺省回退取 key）
func (t pluginTag) DisplayName() string {
	return lo.Ternary(t.name != "", t.name, t.key)
}

// ModelUID 返回关联子模型的 UID（优先取 model，缺省回退取 key）
func (t pluginTag) ModelUID() string {
	return lo.Ternary(t.model != "", t.model, t.key)
}

// parsePluginTag 解析结构体字段上的 `plugin` tag
func parsePluginTag(field reflect.StructField) pluginTag {
	raw := field.Tag.Get("plugin")
	if raw == "-" {
		return pluginTag{skip: true}
	}
	if raw == "" {
		return pluginTag{key: fallbackFieldKey(field)}
	}

	parts := strings.Split(raw, ",")
	tag := pluginTag{}

	// 首项若非键值对且非保留修饰关键字，则作为字段标识键 key
	first := strings.TrimSpace(parts[0])
	if first != "" && !strings.Contains(first, "=") && !isReservedTagKeyword(first) {
		tag.key = first
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if part == "required" {
			tag.required = true
			continue
		}

		if part == "secure" || part == "encrypt" {
			tag.secure = lo.ToPtr(true)
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
		switch k {
		case "secure", "encrypt":
			tag.secure = lo.ToPtr(v == "true" || v == "1")
		case "name", "label", "title", "model_name", "model_title":
			tag.name = v
		case "field":
			tag.field = v
		case "model":
			tag.model = v
		case "group", "model_group":
			tag.group = v
		case "in":
			tag.relationType = v
			tag.direction = types.DirectionToSource
		case "out":
			tag.relationType = v
			tag.direction = types.DirectionToTarget
		case "rel", "relation", "relation_type":
			tag.relationType = v
		case "field_type":
			tag.fieldType = v
		case "type":
			// 兼顾关联关系类型与普通属性字段类型
			tag.relationType = v
			tag.fieldType = v
		case "options", "option", "enums":
			rawOpts := strings.Split(v, "|")
			tag.options = lo.FilterMap(rawOpts, func(opt string, _ int) (string, bool) {
				trimmed := strings.TrimSpace(opt)
				return trimmed, trimmed != ""
			})
		case "direction":
			tag.direction = v
		case "cardinality":
			tag.cardinality = v
		case "default":
			tag.defaultValue = v
		}
	}

	if tag.key == "" {
		tag.key = fallbackFieldKey(field)
	}
	return tag
}

// isReservedTagKeyword 判定是否属于独立的 Tag 布尔修饰关键字
func isReservedTagKeyword(word string) bool {
	switch word {
	case "required", "skip", "-", "secure", "encrypt":
		return true
	default:
		return false
	}
}

// fallbackFieldKey 当未在 plugin tag 显式声明名称时，优先回退为 json tag 名称，否则取首字母小写的 Go 字段名
func fallbackFieldKey(field reflect.StructField) string {
	if raw := field.Tag.Get("json"); raw != "" {
		name := strings.Split(raw, ",")[0]
		if name != "" && name != "-" {
			return name
		}
	}
	return strings.ToLower(field.Name[:1]) + field.Name[1:]
}
