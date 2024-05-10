package domain

type Attribute struct {
	ID           int64
	ModelUid     string
	FieldGroupId int64
	FieldUid     string
	FieldName    string
	FieldType    string
	Required     bool
	Display      bool
	Index        int64
}

type AttributeGroup struct {
	ID    int64
	Name  string
	Index int64
}

// AttributeProjection 映射字段信息
type AttributeProjection struct {
	Projection map[string]int
}
