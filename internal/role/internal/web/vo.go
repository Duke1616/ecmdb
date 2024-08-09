package web

type CreateRoleReq struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Desc   string `json:"desc"`
	Status bool   `json:"status"`
}

type UpdateRoleReq struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Desc   string `json:"desc"`
	Status bool   `json:"status"`
}

type Role struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Desc   string `json:"desc"`
	Status bool   `json:"status"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type RetrieveRoles struct {
	Roles []Role `json:"roles"`
	Total int64  `json:"total"`
}
