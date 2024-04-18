package domain

import "time"

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

type ResourceDiagram struct {
	SRC []ResourceRelation
	DST []ResourceRelation
}
