package domain

import "fmt"

type ResourceRelation struct {
	ID               int64
	SourceModelUID   string
	TargetModelUID   string
	SourceResourceID int64
	TargetResourceID int64
	RelationTypeUID  string // 关联类型唯一索引
	RelationName     string // 拼接字符
}

// ValidateAndComplete 传入对应的模型拓扑关系定义，让资源关系领域对象进行自我校验与补全
func (r *ResourceRelation) ValidateAndComplete(mr ModelRelation) error {
	// 1. 如果当前没有填充 Model 相关的 UID 字段（常见于极简 REST API 入参），则利用注册的拓扑定义自动安全补全
	if r.SourceModelUID == "" {
		r.SourceModelUID = mr.SourceModelUID
		r.RelationTypeUID = mr.RelationTypeUID
		r.TargetModelUID = mr.TargetModelUID
	}

	// 2. 拓扑一致性强校验：确保当前实例的两端模型及关系类型，完全契合注册的拓扑规则
	if r.SourceModelUID != mr.SourceModelUID || r.TargetModelUID != mr.TargetModelUID || r.RelationTypeUID != mr.RelationTypeUID {
		return fmt.Errorf("实例拓扑模型 [%s -> %s (%s)] 与拓扑定义规则 [%s -> %s (%s)] 不一致",
			r.SourceModelUID, r.TargetModelUID, r.RelationTypeUID,
			mr.SourceModelUID, mr.TargetModelUID, mr.RelationTypeUID)
	}

	// 3. 补齐关系实例唯一名称字段以供后续准确检索
	if r.RelationName == "" {
		r.RelationName = mr.RM()
	}
	return nil
}

type ResourceAggregatedAssets struct {
	RelationName string
	ModelUid     string
	Total        int
	ResourceIds  []int64
}

type ResourceDiagram struct {
	SRC []ResourceRelation
	DST []ResourceRelation
}
