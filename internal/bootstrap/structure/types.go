package structure

// Config 表示完整的配置文件结构
type Config struct {
	// ModelGroups 模型分组列表
	ModelGroups []ModelGroupConfig `yaml:"model_groups,omitempty" json:"model_groups,omitempty"`
	// Models 模型列表
	Models []ModelConfig `yaml:"models" json:"models"`
	// RelationTypes 关联类型列表
	RelationTypes []RelationTypeConfig `yaml:"relation_types" json:"relation_types"`
	// ModelRelations 模型关联关系列表
	ModelRelations []ModelRelationConfig `yaml:"model_relations" json:"model_relations"`
}

// ModelGroupConfig 模型分组配置
type ModelGroupConfig struct {
	// Name 分组名称
	Name string `yaml:"name" json:"name"`
}

// ModelConfig 模型配置
type ModelConfig struct {
	// UID 模型唯一标识
	UID string `yaml:"uid" json:"uid"`
	// Name 模型名称
	Name string `yaml:"name" json:"name"`
	// Icon 模型图标
	Icon string `yaml:"icon" json:"icon"`
	// GroupName 模型分组名称（可选）
	GroupName string `yaml:"group_name,omitempty" json:"group_name,omitempty"`
	// Builtin 是否为内置模型
	Builtin bool `yaml:"builtin" json:"builtin"`
	// Attributes 属性配置
	Attributes AttributesConfig `yaml:"attributes" json:"attributes"`
}

// AttributesConfig 属性配置（包含分组和字段）
type AttributesConfig struct {
	// Groups 属性分组列表
	Groups []AttributeGroupConfig `yaml:"groups" json:"groups"`
}

// AttributeGroupConfig 属性分组配置
type AttributeGroupConfig struct {
	// Name 分组名称
	Name string `yaml:"name" json:"name"`
	// Index 分组排序索引
	Index int64 `yaml:"index" json:"index"`
	// Fields 字段列表
	Fields []FieldConfig `yaml:"fields" json:"fields"`
}

// FieldConfig 字段配置
type FieldConfig struct {
	// UID 字段唯一标识
	UID string `yaml:"uid" json:"uid"`
	// Name 字段名称
	Name string `yaml:"name" json:"name"`
	// Type 字段类型 (string, number, list, text, multiline 等)
	Type string `yaml:"type" json:"type"`
	// Option 字段选项（用于 list 类型）
	Option []string `yaml:"option,omitempty" json:"option,omitempty"`
	// Required 是否必填
	Required bool `yaml:"required" json:"required"`
	// Display 是否在列表中显示
	Display bool `yaml:"display" json:"display"`
	// Secure 是否为加密字段
	Secure bool `yaml:"secure" json:"secure"`
	// Builtin 是否为内置字段
	Builtin bool `yaml:"builtin" json:"builtin"`
	// Index 字段排序索引
	Index int64 `yaml:"index" json:"index"`
}

// RelationTypeConfig 关联类型配置
type RelationTypeConfig struct {
	// UID 关联类型唯一标识
	UID string `yaml:"uid" json:"uid"`
	// Name 关联类型名称
	Name string `yaml:"name" json:"name"`
	// SourceDescribe 源端描述
	SourceDescribe string `yaml:"source_describe" json:"source_describe"`
	// TargetDescribe 目标端描述
	TargetDescribe string `yaml:"target_describe" json:"target_describe"`
}

// ModelRelationConfig 模型关联关系配置
type ModelRelationConfig struct {
	// SourceModelUID 源模型 UID
	SourceModelUID string `yaml:"source_model_uid" json:"source_model_uid"`
	// TargetModelUID 目标模型 UID
	TargetModelUID string `yaml:"target_model_uid" json:"target_model_uid"`
	// RelationTypeUID 关联类型 UID
	RelationTypeUID string `yaml:"relation_type_uid" json:"relation_type_uid"`
	// RelationName 关联关系名称
	RelationName string `yaml:"relation_name" json:"relation_name"`
	// Mapping 映射类型 (one_to_one, one_to_many, many_to_many)
	Mapping string `yaml:"mapping" json:"mapping"`
}
