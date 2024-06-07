package web

import (
	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
)

type CreateAttributeReq struct {
	GroupId   int64       `json:"group_id"`
	FieldUid  string      `json:"field_uid"`
	ModelUid  string      `json:"model_uid"`
	FieldName string      `json:"field_name"`
	FieldType string      `json:"field_type"`
	Secure    bool        `json:"secure"`
	Required  bool        `json:"required"`
	Option    interface{} `json:"option"`
}

type CreateAttributeGroup struct {
	Name     string `json:"group_name"`
	ModelUid string `json:"model_uid"`
	Index    int64  `json:"index"`
}

type ListAttributeGroupReq struct {
	ModelUid string `json:"model_uid"`
}
type ListAttributeGroupByIdsReq struct {
	Ids []int64 `json:"ids"`
}

type ListAttributeReq struct {
	ModelUid string `json:"model_uid"`
}

type DeleteAttributeReq struct {
	Id int64 `json:"id"`
}

type Attribute struct {
	ID        int64       `json:"id"`
	ModelUid  string      `json:"model_uid"`
	FieldUid  string      `json:"field_uid"`
	FieldName string      `json:"field_name"`
	FieldType string      `json:"field_type"`
	Required  bool        `json:"required"`
	Secure    bool        `json:"secure"`
	Display   bool        `json:"display"`
	Option    interface{} `json:"option"`
	Index     int64       `json:"index"`
}

type AttributeGroup struct {
	GroupName string `json:"group_name"`
	ModelUid  string `json:"model_uid"`
	GroupId   int64  `json:"group_id"`
	Index     int64  `json:"index"`
}

// CustomAttributeFieldColumnsReq 排序并展示数据
type CustomAttributeFieldColumnsReq struct {
	ModelUid        string   `json:"model_uid"`
	CustomFieldName []string `json:"custom_field_name"`
}

type RetrieveAttributeFieldsList struct {
}

type AttributeGroups struct {
	GroupId    int64   `bson:"_id"`
	Total      int     `bson:"total"`
	Attributes []int64 `bson:"attributes"`
}

type AttributeList struct {
	GroupId    int64       `json:"group_id"`
	GroupName  string      `json:"group_name"`
	Expanded   bool        `json:"expanded"`
	Index      int64       `json:"index"`
	Total      int         `json:"total"`
	Attributes []Attribute `json:"attributes,omitempty"`
}

type RetrieveAttributeList struct {
	AttributeList []AttributeList `json:"attribute_groups"`
}

type RetrieveAttributeFieldList struct {
	Attributes []Attribute `json:"attribute_fields,omitempty"`
	Total      int64       `json:"total,omitempty"`
}

func toDomain(req CreateAttributeReq) domain.Attribute {
	return domain.Attribute{
		GroupId:   req.GroupId,
		FieldUid:  req.FieldUid,
		ModelUid:  req.ModelUid,
		FieldName: req.FieldName,
		FieldType: req.FieldType,
		Required:  req.Required,
		Secure:    req.Secure,
		Option:    req.Option,
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
		Option:    attr.Option,
		Secure:    attr.Secure,
		Index:     attr.Index,
	}
}
