package domain

type ResourceRelation struct {
	ID               int64
	SourceModelUID   string
	TargetModelUID   string
	SourceResourceID int64
	TargetResourceID int64
	RelationTypeUID  string // 关联类型唯一索引
	RelationName     string // 拼接字符
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
