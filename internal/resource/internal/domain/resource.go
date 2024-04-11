package domain

import "github.com/Duke1616/ecmdb/pkg/mongox"

type Resource struct {
	ID              int64
	ModelIdentifies string // 因为这个传参是 URL PATH, 使用ID会显得丑陋，所以使用模型唯一身份标识
	Data            mongox.MapStr
}

type DetailResource struct {
	ID         int64       `json:"id"`
	Attributes []Attribute `json:"attributes,omitempty"`
}

type Attribute struct {
	ID              int64
	ModelIdentifies string
	Identifies      string
	Name            string
	FieldType       string
	Required        bool
}
