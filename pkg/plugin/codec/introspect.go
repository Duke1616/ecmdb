package codec

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
)

// IEnum 允许属性字段类型通过实现该接口自描述其合法枚举选项列表
type IEnum interface {
	EnumOptions() []string
}

// CompactModelDescriptor 允许模型结构体通过单方法自描述中文名称与所属模型分组
type CompactModelDescriptor interface {
	DescribeModel() (name string, group string)
}

// ModelDescriptor 允许模型结构体自描述其中文名称与所属模型分组
type ModelDescriptor interface {
	ModelName() string
	ModelGroup() string
}

// ModelNameDescriptor 允许模型结构体仅自描述其中文展示名称
type ModelNameDescriptor interface {
	ModelName() string
}

// ModelGroupDescriptor 允许模型结构体仅自描述其所属模型分组
type ModelGroupDescriptor interface {
	ModelGroup() string
}

// TargetMeta 包含目标资产经过一次内省推导产出的 Schema 和 ResourceSpec
type TargetMeta struct {
	Schema types.Schema
	Spec   types.ResourceSpec
}

// InspectTarget 基于目标资产泛型 T 完成单次内省与双向投影，零冗余分支
func InspectTarget[T any](name string, modelUID string) (TargetMeta, error) {
	if name == "" {
		name = "target"
	}
	var zero T
	node, err := inspectStructNode(reflect.TypeOf(zero), name, modelUID, pluginTag{required: true})
	if err != nil {
		return TargetMeta{}, err
	}

	return TargetMeta{
		Schema: node.toSchema(),
		Spec:   node.toResourceSpec(),
	}, nil
}

// structField 保存模型中解析出的普通属性字段元数据
type structField struct {
	Key         string   // Go struct 字段标识（用于 Spec.Fields 映射键）
	CMDBUID     string   // CMDB 属性字段 UID（用于 Spec.Fields 映射值与 Attribute.UID）
	DisplayName string   // 前端展示名称（name 优先，回退到 Key）
	Type        string   // CMDB 属性类型（string, list, int, float, bool 等）
	Options     []string // 枚举下拉选项列表
	Required    bool     // 是否必填
	Secure      bool     // 是否属于敏感加密字段
}

// structNode 代表经过一次内省反射分析后的结构体节点元数据树
type structNode struct {
	Name         string        // 资源名称 (Go struct 中的字段名)
	ModelUID     string        // CMDB 模型 UID
	ModelName    string        // CMDB 模型中文名称
	GroupName    string        // 所属 CMDB 模型分组名称
	RelationType string        // 关联类型 UID
	Direction    string        // 关联查询方向
	Cardinality  string        // 数量基数 (one / many)
	Required     bool          // 是否必填
	fields       []structField // 本模型的所有普通属性字段
	Children     []*structNode // 递归子关联模型列表
}

// inspectStructNode 对给定类型 t 执行单次深度内省，构建 structNode 元数据树
func inspectStructNode(t reflect.Type, name string, modelUID string, tag pluginTag) (*structNode, error) {
	t = peelType(t)
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("target must be a struct, got %s", t.Kind())
	}

	if modelUID == "" {
		modelUID = tag.ModelUID()
	}

	node := &structNode{
		Name:         name,
		ModelUID:     strings.TrimSpace(modelUID),
		ModelName:    resolveModelName(t, tag),
		GroupName:    resolveGroupName(t, tag),
		RelationType: tag.relationType,
		Direction:    tag.direction,
		Cardinality:  lo.Ternary(tag.cardinality != "", tag.cardinality, types.CardinalityOne),
		Required:     tag.required,
	}

	if node.RelationType != "" && !types.ValidRelationType(node.RelationType) {
		return nil, fmt.Errorf("unsupported relation type %s", node.RelationType)
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}

		fieldTag := parsePluginTag(field)
		if fieldTag.skip {
			continue
		}

		if field.Anonymous {
			if err := node.mergeEmbedded(field, fieldTag); err != nil {
				return nil, err
			}
			continue
		}

		if isStruct, elemType := isStructOrStructSlice(field.Type); isStruct {
			if err := node.addChild(elemType, field.Type, field.Name, fieldTag); err != nil {
				return nil, err
			}
		} else {
			node.addField(fieldTag, field.Type)
		}
	}

	return node, nil
}

func (n *structNode) mergeEmbedded(field reflect.StructField, tag pluginTag) error {
	underlying := peelType(field.Type)
	if underlying.Kind() != reflect.Struct {
		return nil
	}
	childNode, err := inspectStructNode(underlying, "", "", tag)
	if err != nil {
		return err
	}
	if tag.name != "" {
		n.ModelName = tag.name
	} else if n.ModelName == "" {
		n.ModelName = childNode.ModelName
	}

	if tag.group != "" {
		n.GroupName = tag.group
	} else if n.GroupName == "" {
		n.GroupName = childNode.GroupName
	}
	n.fields = append(n.fields, childNode.fields...)
	n.Children = append(n.Children, childNode.Children...)
	return nil
}

func (n *structNode) addChild(elemType, rawType reflect.Type, fieldName string, tag pluginTag) error {
	if tag.cardinality == "" {
		rawElem := peelType(rawType)
		tag.cardinality = lo.Ternary(rawElem.Kind() == reflect.Slice || rawElem.Kind() == reflect.Array,
			types.CardinalityMany, types.CardinalityOne)
	}

	childNode, err := inspectStructNode(elemType, tag.key, tag.ModelUID(), tag)
	if err != nil {
		return fmt.Errorf("inspect child %s error: %w", fieldName, err)
	}
	n.Children = append(n.Children, childNode)
	return nil
}

// extractEnumOptions 从反射类型中探测 IEnum 接口，兼容值接收者与指针接收者
func extractEnumOptions(t reflect.Type) ([]string, bool) {
	if t.Kind() == reflect.Interface {
		return nil, false
	}
	if enum, ok := reflect.Zero(t).Interface().(IEnum); ok {
		return enum.EnumOptions(), true
	}
	if enum, ok := reflect.New(t).Interface().(IEnum); ok {
		return enum.EnumOptions(), true
	}
	return nil, false
}

// resolveFieldTypeAndOptions 解析最终落库的 CMDB 字段类型与枚举选项
func resolveFieldTypeAndOptions(tag pluginTag, rawType reflect.Type) (string, []string) {
	// 优先策略 1：实现了 IEnum 自描述接口的强类型枚举，自动推导为 list
	if opts, ok := extractEnumOptions(rawType); ok {
		return "list", opts
	}
	// 优先策略 2：Tag 中通过 options= 显式声明了候选项
	if len(tag.options) > 0 {
		return lo.Ternary(tag.fieldType != "", tag.fieldType, "list"), tag.options
	}
	// 优先策略 3：Tag 中显式指定了 field_type/type（如 multiline）
	if tag.fieldType != "" {
		return tag.fieldType, nil
	}
	// 兜底策略 4：基于 Go 原生类型智能推断 (datetime, multiline, boolean, number, string)
	return inferAttributeType(rawType), nil
}

func (n *structNode) addField(tag pluginTag, t reflect.Type) {
	cmdbUID := tag.CMDBUID()
	rawType := peelType(t)
	fieldType, options := resolveFieldTypeAndOptions(tag, rawType)

	n.fields = append(n.fields, structField{
		Key:         tag.key,
		CMDBUID:     cmdbUID,
		DisplayName: tag.DisplayName(),
		Type:        fieldType,
		Options:     options,
		Required:    tag.required,
		Secure:      isSecureField(cmdbUID, tag.key),
	})
}

// toResourceSpec 投影为 Binding Graph 使用的 types.ResourceSpec
func (n *structNode) toResourceSpec() types.ResourceSpec {
	spec := types.ResourceSpec{
		Name:         n.Name,
		ModelUID:     n.ModelUID,
		ModelName:    n.ModelName,
		GroupName:    n.GroupName,
		RelationType: n.RelationType,
		Direction:    n.Direction,
		Cardinality:  n.Cardinality,
		Required:     n.Required,
		Fields:       make(map[string]string, len(n.fields)),
	}

	for _, f := range n.fields {
		spec.Fields[f.Key] = f.CMDBUID
		if f.Required {
			spec.RequiredFields = append(spec.RequiredFields, f.Key)
		}
	}

	spec.Children = lo.Map(n.Children, func(c *structNode, _ int) types.ResourceSpec {
		return c.toResourceSpec()
	})

	return spec
}

// toSchema 将内省树投影为完整的 CMDB Schema（模型、属性组、关联关系）
func (n *structNode) toSchema() types.Schema {
	schema := types.Schema{
		RelationTypes: types.BasicRelationTypes(),
	}
	models := make(map[string]types.ModelSpec)
	groups := make(map[string]struct{})

	n.collectInto(&schema, models, groups)

	schema.Models = lo.Values(models)
	schema.ModelGroups = lo.Map(lo.Keys(groups), func(name string, _ int) types.ModelGroupSpec {
		return types.ModelGroupSpec{Name: name}
	})
	return schema
}

func (n *structNode) collectInto(schema *types.Schema, models map[string]types.ModelSpec, groups map[string]struct{}) {
	modelUID := strings.TrimSpace(n.ModelUID)
	if modelUID == "" {
		return
	}

	if _, exists := models[modelUID]; !exists {
		if n.GroupName != "" {
			groups[n.GroupName] = struct{}{}
		}
		models[modelUID] = n.toModelSpec()
	}

	for _, child := range n.Children {
		if child.RelationType != "" && child.ModelUID != "" {
			n.appendRelation(schema, child)
		}
		child.collectInto(schema, models, groups)
	}
}

func (n *structNode) toModelSpec() types.ModelSpec {
	name := lo.Ternary(n.ModelName != "", n.ModelName, n.ModelUID)

	attrs := lo.Map(n.fields, func(f structField, idx int) types.Attribute {
		return types.Attribute{
			UID:      f.CMDBUID,
			Name:     f.DisplayName,
			Type:     f.Type,
			Option:   f.Options,
			Required: f.Required,
			Display:  !f.Secure,
			Secure:   f.Secure,
			Builtin:  true,
			Index:    int64(idx + 1),
		}
	})

	return types.ModelSpec{
		UID:       n.ModelUID,
		Name:      name,
		GroupName: n.GroupName,
		Builtin:   true,
		AttributeGroups: []types.AttributeGroup{
			{
				Name:   "基础属性",
				Index:  0,
				Fields: attrs,
			},
		},
	}
}

func (n *structNode) appendRelation(schema *types.Schema, child *structNode) {
	rel := types.ModelRelation{
		RelationTypeUID: child.RelationType,
		Mapping:         types.MappingManyToMany,
	}
	if child.Direction == types.DirectionToSource {
		rel.SourceModelUID = child.ModelUID
		rel.TargetModelUID = n.ModelUID
	} else {
		rel.SourceModelUID = n.ModelUID
		rel.TargetModelUID = child.ModelUID
	}

	exists := lo.SomeBy(schema.ModelRelations, func(r types.ModelRelation) bool {
		return r.SourceModelUID == rel.SourceModelUID &&
			r.TargetModelUID == rel.TargetModelUID &&
			r.RelationTypeUID == rel.RelationTypeUID
	})
	if !exists {
		schema.ModelRelations = append(schema.ModelRelations, rel)
	}
}

func resolveModelName(t reflect.Type, tag pluginTag) string {
	if name := extractFromDescriptor[CompactModelDescriptor](t, func(d CompactModelDescriptor) string {
		name, _ := d.DescribeModel()
		return name
	}); name != "" {
		return name
	}
	if name := extractFromDescriptor[ModelNameDescriptor](t, func(d ModelNameDescriptor) string { return d.ModelName() }); name != "" {
		return name
	}
	return tag.DisplayName()
}

func resolveGroupName(t reflect.Type, tag pluginTag) string {
	if group := extractFromDescriptor[CompactModelDescriptor](t, func(d CompactModelDescriptor) string {
		_, group := d.DescribeModel()
		return group
	}); group != "" {
		return group
	}
	if group := extractFromDescriptor[ModelGroupDescriptor](t, func(d ModelGroupDescriptor) string { return d.ModelGroup() }); group != "" {
		return group
	}
	return tag.group
}

func extractFromDescriptor[D any](t reflect.Type, getter func(D) string) string {
	if t.Kind() != reflect.Struct {
		return ""
	}
	if d, ok := reflect.Zero(t).Interface().(D); ok {
		return getter(d)
	}
	if d, ok := reflect.New(t).Interface().(D); ok {
		return getter(d)
	}
	return ""
}

// isTimeType 判定反射类型是否为 time.Time
func isTimeType(t reflect.Type) bool {
	return t.PkgPath() == "time" && t.Name() == "Time"
}

// isByteSlice 判定反射类型是否为 []byte
func isByteSlice(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8
}

// inferAttributeType 将 Go 反射类型智能映射为 CMDB 标准属性字段数据类型
func inferAttributeType(t reflect.Type) string {
	t = peelType(t)

	if isTimeType(t) {
		return "datetime"
	}
	if isByteSlice(t) {
		return "multiline"
	}

	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	default:
		return "string"
	}
}

// isSecureField 判定字段是否属于需异步加密的安全敏感字段
func isSecureField(fieldName, tagName string) bool {
	lower := strings.ToLower(fieldName + " " + tagName)
	secureKeywords := []string{"password", "private_key", "secret", "token"}
	return lo.SomeBy(secureKeywords, func(kw string) bool {
		return strings.Contains(lower, kw)
	})
}
