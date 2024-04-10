package web

type MapStr map[string]interface{}

type CreateResourceReq struct {
	Data MapStr `json:"data"`
}
