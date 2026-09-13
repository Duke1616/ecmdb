package types

import "github.com/samber/lo"

// ── 操作栏位置 ────────────────────────────────────────────────────────────────

const PlacementResourceDetailActions = "resource.detail.actions"

// ── 运行时模式与通信协议常量 ──────────────────────────────────────────────────

const (
	RuntimeModeBuiltin         = "builtin"
	RuntimeModeExternalService = "external-service"

	HeaderPluginID = "X-ECMDB-Plugin-ID"
	WellKnownPath  = "/.well-known/ecmdb-plugin"
)

// ── 关联查询方向 ──────────────────────────────────────────────────────────────

const (
	DirectionToSource = "source"
	DirectionToTarget = "target"
)

// ── 关联类型 UID ───────────────────────────────────────────────────────────────

const (
	RelationTypeDefault = "default"
	RelationTypeGroup   = "group"
	RelationTypeBelong  = "belong"
	RelationTypeRun     = "run"
)

// ── 关联映射基数 ──────────────────────────────────────────────────────────────

const (
	MappingOneToOne   = "one_to_one"
	MappingOneToMany  = "one_to_many"
	MappingManyToMany = "many_to_many"
)

// ── 节点基数（Cardinality）────────────────────────────────────────────────────

const (
	CardinalityOne  = "one"
	CardinalityMany = "many"
)

// validRelationTypes 合法关联类型枚举集合
var validRelationTypes = []string{
	RelationTypeDefault,
	RelationTypeGroup,
	RelationTypeBelong,
	RelationTypeRun,
}

// ValidRelationType 校验 relationType 是否为平台已知关联类型
func ValidRelationType(relationType string) bool {
	return lo.Contains(validRelationTypes, relationType)
}

// BasicRelationTypes 返回 ECMDB 内置预设的基础关联类型定义
func BasicRelationTypes() []RelationType {
	return []RelationType{
		{
			UID:            RelationTypeDefault,
			Name:           "默认关联",
			SourceDescribe: "关联",
			TargetDescribe: "关联",
		},
		{
			UID:            RelationTypeRun,
			Name:           "运行",
			SourceDescribe: "运行于",
			TargetDescribe: "运行",
		},
		{
			UID:            RelationTypeGroup,
			Name:           "组成",
			SourceDescribe: "组成",
			TargetDescribe: "组成于",
		},
		{
			UID:            RelationTypeBelong,
			Name:           "属于",
			SourceDescribe: "属于",
			TargetDescribe: "包含",
		},
	}
}

