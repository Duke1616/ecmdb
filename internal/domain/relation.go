package domain

import (
	"fmt"
	"time"
)

const (
	MappingOneToOne   = iota + 1 // 一对一关系
	MappingOneToMany             // 一对多关系
	MappingManyToMany            // 多对多关系
)

type ModelRelation struct {
	ID              int64
	SourceModelUID  string
	TargetModelUID  string
	RelationTypeUID string // 关联类型唯一索引
	RelationName    string // 拼接字符
	Mapping         string // 关联关系
	Ctime           time.Time
	Utime           time.Time
}

func (m *ModelRelation) RM() string {
	return fmt.Sprintf("%s_%s_%s", m.SourceModelUID, m.RelationTypeUID, m.TargetModelUID)
}

// IsSource 判定指定的 modelUid 是否为该关系定义的源端
func (m ModelRelation) IsSource(modelUid string) bool {
	// NOTE: 在拓扑定义中，SourceModelUID 表示关联关系的源端模型唯一标识
	return m.SourceModelUID == modelUid
}

// IsTarget 判定指定的 modelUid 是否为该关系定义的目标端
func (m ModelRelation) IsTarget(modelUid string) bool {
	// NOTE: 在拓扑定义中，TargetModelUID 表示关联关系的目标端模型唯一标识
	return m.TargetModelUID == modelUid
}

// BelongsTo 判定指定的 modelUid 是否参与了该关系定义
func (m ModelRelation) BelongsTo(modelUid string) bool {
	// NOTE: 校验模型是否属于该关联的两端（源端或目标端）
	return m.SourceModelUID == modelUid || m.TargetModelUID == modelUid
}

// Validate 校验模型关联定义本身的合法性并自我补齐关联名
func (m *ModelRelation) Validate() error {
	// NOTE: 校验源模型、目标模型及关系类型定义是否完整
	if m.SourceModelUID == "" || m.TargetModelUID == "" || m.RelationTypeUID == "" {
		return fmt.Errorf("模型关联的源端、目标端或关系类型 UID 不能为空")
	}
	// 自动完成 RelationName 的一致性生成与补齐，提供强一致性保障
	m.RelationName = m.RM()
	return nil
}

// ModelDiagram 拓补图模型关联节点信息
type ModelDiagram struct {
	ID              int64
	RelationTypeUid string
	TargetModelUid  string
	SourceModelUid  string
}

type RelationType struct {
	ID             int64
	Name           string
	UID            string
	SourceDescribe string
	TargetDescribe string
	Ctime          time.Time
	Utime          time.Time
}

// Validate 校验关联类型定义合法性
func (r *RelationType) Validate() error {
	// NOTE: 检查基本的名字及唯一索引 UID 标识是否存在
	if r.Name == "" || r.UID == "" {
		return fmt.Errorf("关联类型的 Name 和 UID 不能为空")
	}
	return nil
}
