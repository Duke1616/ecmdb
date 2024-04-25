package domain

type Attribute struct {
	ID        int64
	ModelUID  string
	Name      string
	FieldName string
	FieldType string
	Required  bool
}

// AttributeProjection 映射字段信息
type AttributeProjection struct {
	Projection map[string]int
}
