package web

import (
	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
)

type CreateAttributeReq struct {
	FieldUid  string `json:"field_uid"`
	ModelUid  string `json:"model_uid"`
	FieldName string `json:"field_name"`
	FieldType string `json:"field_type"`
	Required  bool   `json:"required"`
}

type CreateAttributeGroup struct {
	Name string `json:"name"`
}

type ListAttributeGroupReq struct {
	ModelUid string `json:"model_uid"`
}

type ListAttributeReq struct {
	ModelUid string `json:"model_uid"`
}

type DeleteAttributeReq struct {
	Id int64 `json:"id"`
}

type Attribute struct {
	ID        int64  `json:"id"`
	ModelUid  string `json:"model_uid"`
	FieldUid  string `json:"field_uid"`
	FieldName string `json:"field_name"`
	FieldType string `json:"field_type"`
	Required  bool   `json:"required"`
	Display   bool   `json:"display"`
	Index     int64  `json:"index"`
}

type AttributeGroup struct {
	GroupName  string      `json:"group_name"`
	Expanded   bool        `json:"expanded"`
	GroupId    int64       `json:"group_id"`
	Attributes []Attribute `json:"attributes"`
}

// CustomAttributeFieldColumnsReq 排序并展示数据
type CustomAttributeFieldColumnsReq struct {
	ModelUid        string   `json:"model_uid"`
	CustomFieldName []string `json:"custom_field_name"`
}

type RetrieveAttributeFieldsList struct {
}

type RetrieveAttributeList struct {
	Attributes []AttributeGroup `json:"ags,omitempty"`
	Total      int64            `json:"total,omitempty"`
}

type RetrieveAttributeFieldList struct {
	Attributes []Attribute `json:"attribute_fields,omitempty"`
	Total      int64       `json:"total,omitempty"`
}

func toDomain(req CreateAttributeReq) domain.Attribute {
	return domain.Attribute{
		FieldUid:  req.FieldUid,
		ModelUid:  req.ModelUid,
		FieldName: req.FieldName,
		FieldType: req.FieldType,
		Required:  req.Required,
	}
}

func toAttributeVo(attr domain.Attribute) Attribute {
	return Attribute{
		ID:        attr.ID,
		FieldUid:  attr.FieldUid,
		ModelUid:  attr.ModelUid,
		FieldName: attr.FieldName,
		FieldType: attr.FieldType,
		Required:  attr.Required,
		Display:   attr.Display,
		Index:     attr.Index,
	}
}
