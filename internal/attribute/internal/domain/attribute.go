package domain

type Attribute struct {
	ID        int64
	GroupId   int64
	ModelUid  string
	FieldUid  string
	FieldName string
	FieldType string
	Required  bool
	Display   bool
	Secure    bool
	Link      bool
	Index     int64
	Option    interface{}
	Builtin   bool
}

type AttributeGroup struct {
	ID       int64
	Name     string
	ModelUid string
	Index    int64
}

type AttributePipeline struct {
	GroupId    int64       `bson:"_id"`
	Total      int         `bson:"total"`
	Attributes []Attribute `bson:"attributes"`
}
