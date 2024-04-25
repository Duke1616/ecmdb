package web

import (
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"time"
)

type CreateModelGroupReq struct {
	Name string `json:"name"`
}

type CreateModelReq struct {
	Name    string `json:"name"`
	GroupId int64  `json:"group_id"`
	UID     string `json:"uid"`
	Icon    string `json:"icon"`
}

type DetailModelReq struct {
	ID int64 `json:"id"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type RetrieveModelsListResp struct {
	Total  int64   `json:"total,omitempty"`
	Models []Model `json:"models,omitempty"`
}

type Model struct {
	Name  string `json:"name"`
	UID   string `json:"uid"`
	Icon  string `json:"icon"`
	Ctime string `json:"ctime"`
	Utime string `json:"utime"`
}

func toModelVo(m domain.Model) Model {
	return Model{
		Name:  m.Name,
		UID:   m.UID,
		Ctime: m.Utime.Format(time.DateTime),
		Utime: m.Utime.Format(time.DateTime),
	}
}
