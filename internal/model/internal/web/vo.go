package web

type CreateModelGroupReq struct {
	Name string `json:"name"`
}

type CreateModelReq struct {
	Name    string `json:"name"`
	GroupId int64  `json:"group_id"`
	UID     string `json:"uid"`
}

type DetailUidModelReq struct {
	uid string `json:"uid"`
}

type ListModelsReq struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListModelsResp struct {
	Total  int64   `json:"total,omitempty"`
	Models []Model `json:"models,omitempty"`
}

type Model struct {
	Name  string `json:"name"`
	UID   string `json:"uid"`
	Ctime string `json:"ctime"`
	Utime string `json:"utime"`
}
