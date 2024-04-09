package web

type CreateModelGroupReq struct {
	Name             string `json:"name"`
	UniqueIdentifier string `json:"unique_identifier"`
}

type CreateModelReq struct {
	Name string `json:"name"`
}

type CreateModelAttrReq struct {
	ModelID int64 `json:"model_id"`
}
