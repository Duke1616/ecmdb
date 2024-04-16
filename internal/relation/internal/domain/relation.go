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

type ResourceRelation struct {
	ID               int64
	SourceModelUID   string
	TargetModelUID   string
	SourceResourceID int64
	TargetResourceID int64
	RelationTypeUID  string // 关联类型唯一索引
	RelationName     string // 拼接字符
	Ctime            time.Time
	Utime            time.Time
}

type RelationType struct {
	ID             int64
	UID            string
	SourceDescribe string
	TargetDescribe string
	Ctime          time.Time
	Utime          time.Time
}
