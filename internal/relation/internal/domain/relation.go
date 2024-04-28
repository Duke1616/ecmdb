package domain

import (
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
