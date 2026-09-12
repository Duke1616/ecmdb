package domain

import (
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/samber/lo"
)

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

// AdminSearchModelCount DAO 聚合纯计数结果
type AdminSearchModelCount struct {
	TenantID int64
	ModelUID string
	Total    int
}

// AdminSearchStructureModel 大盘模型维度统计
type AdminSearchStructureModel struct {
	ModelUID  string `json:"model_uid"`
	ModelName string `json:"model_name"`
	Total     int    `json:"total"`
}

// AdminSearchStructureTenant 大盘租户维度统计与模型 Tabs
type AdminSearchStructureTenant struct {
	TenantID   int64                       `json:"tenant_id"`
	TenantType string                      `json:"tenant_type"`
	Total      int                         `json:"total"`
	Models     []AdminSearchStructureModel `json:"models"`
}

// AdminSearchStructureResult 全局大盘结构与导航响应
type AdminSearchStructureResult struct {
	Total         int                          `json:"total"`
	Organizations []AdminSearchStructureTenant `json:"organizations"`
	Personals     []AdminSearchStructureTenant `json:"personals"`
}

// ModelUIDs 提取跨租户大盘中所有命中的唯一模型标识
func (r AdminSearchStructureResult) ModelUIDs() []string {
	return lo.Uniq(lo.FlatMap(append(r.Organizations, r.Personals...), func(t AdminSearchStructureTenant, _ int) []string {
		return lo.Map(t.Models, func(m AdminSearchStructureModel, _ int) string {
			return m.ModelUID
		})
	}))
}

// SearchStructureResult 单租户检索模型 Tabs 概览
type SearchStructureResult struct {
	Total  int                         `json:"total"`
	Models []AdminSearchStructureModel `json:"models"`
}

// ModelUIDs 提取单租户检索命中所有模型的唯一标识
func (r SearchStructureResult) ModelUIDs() []string {
	return lo.Map(r.Models, func(m AdminSearchStructureModel, _ int) string {
		return m.ModelUID
	})
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
