package web

type CreateModelGroupReq struct {
	Name             string `json:"name"`
	UniqueIdentifier string `json:"unique_identifier"`
}

type CreateModelReq struct {
	Name       string `json:"name"`
	GroupId    int64  `json:"group_id"`
	Identifies string `json:"identifies"`
}
