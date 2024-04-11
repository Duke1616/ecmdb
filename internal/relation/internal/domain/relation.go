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
	ID                     int64
	SourceModelIdentifies  string
	TargetModelIdentifies  string
	RelationTypeIdentifies string
	RelationName           string
	Mapping                string
	Ctime                  time.Time
	Utime                  time.Time
}

type ResourceRelation struct {
}
