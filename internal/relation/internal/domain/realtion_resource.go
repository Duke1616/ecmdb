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

type ResourceAggregatedData struct {
	RelationName string
	ModelUid     string
	Count        int
	Data         []ResourceRelation
}

type ResourceDiagram struct {
	SRC []ResourceRelation
	DST []ResourceRelation
}
