package web

type CreateModelGroupReq struct {
	Name string `json:"name"`
}

type CreateModelReq struct {
	Name       string `json:"name"`
	GroupId    int64  `json:"group_id"`
	Identifies string `json:"identifies"`
}

type DetailUniqueIdentifierModelReq struct {
	UniqueIdentifier string `json:"identifies"`
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
	Name       string `json:"name"`
	Identifies string `json:"identifies"`
	Ctime      string `json:"ctime"`
	Utime      string `json:"utime"`
}
