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
