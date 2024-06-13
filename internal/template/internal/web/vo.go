package web

type CreateTemplateReq struct {
	Name    string `json:"name"`
	Rules   string `json:"rules"`
	Options string `json:"options"`
}

type DetailTemplateReq struct {
	Id int64 `json:"id"`
}
