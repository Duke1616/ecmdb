package web

import (
	"github.com/Duke1616/ecmdb/internal/attribute/internal/domain"
)

type CreateAttributeReq struct {
	Name      string `json:"name"`
	ModelUid  string `json:"model_uid"`
	FieldName string `json:"field_name"`
	FieldType string `json:"field_type"`
	Required  bool   `json:"required"`
}

type DetailAttributeReq struct {
	ModelUid string `json:"model_uid"`
}

type ListAttributeReq struct {
	ModelUid string `json:"model_uid"`
}

type Attribute struct {
	ID        int64  `json:"id"`
	ModelUID  string `json:"model_uid"`
	Name      string `json:"name"`
	FieldName string `json:"field_name"`
	FieldType string `json:"field_type"`
	Required  bool   `json:"required"`
}

type RetrieveAttributeList struct {
	Attribute []Attribute `json:"attribute,omitempty"`
	Total     int64       `json:"total,omitempty"`
}

func toDomain(req CreateAttributeReq) domain.Attribute {
	return domain.Attribute{
		Name:      req.Name,
		ModelUID:  req.ModelUid,
		FieldName: req.FieldName,
		FieldType: req.FieldType,
		Required:  req.Required,
	}
}

func toAttributeVo(attr domain.Attribute) Attribute {
	return Attribute{
		ID:        attr.ID,
		Name:      attr.Name,
		ModelUID:  attr.ModelUID,
		FieldName: attr.FieldName,
		FieldType: attr.FieldType,
		Required:  attr.Required,
	}
}
