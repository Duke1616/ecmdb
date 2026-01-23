package domain

import "github.com/Duke1616/ecmdb/pkg/mongox"

// Operator 导出操作符枚举
type Operator string

const (
	OperatorEq       Operator = "eq"
	OperatorNe       Operator = "ne"
	OperatorContains Operator = "contains"
	OperatorGt       Operator = "gt"
	OperatorLt       Operator = "lt"
)

type Resource struct {
	ID       int64         `json:"id"`
	Name     string        `json:"name"`
	ModelUID string        `json:"model_uid"`
	Data     mongox.MapStr `json:"data"`
}

type ResourceRelation struct {
	ModelUid  string
	Resources []Resource
}

type SearchResource struct {
	ModelUid string
	Total    int `json:"total"`
	Data     []mongox.MapStr
}

type Condition struct {
	Name      string `json:"name"`      // 过滤名称
	Condition string `json:"condition"` // 过滤条件
	Input     string `json:"input"`     // 过滤输入
}

// FilterCondition 筛选条件
type FilterCondition struct {
	FieldUID string      `json:"field_uid"`
	Operator Operator    `json:"operator"`
	Value    interface{} `json:"value"`
}

// FilterGroup 筛选条件组 (组内 AND)
type FilterGroup struct {
	Filters []FilterCondition `json:"filters"`
}
