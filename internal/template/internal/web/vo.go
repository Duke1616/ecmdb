package web

type CreateTemplateReq struct {
	Name    string `json:"name"`
	Rules   string `json:"rules"`
	Options string `json:"options"`
	Desc    string `json:"desc"`
}

type DetailTemplateReq struct {
	Id int64 `json:"id"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListTemplateReq struct {
	Page
}

type DeleteTemplateReq struct {
	Id int64 `json:"id"`
}

type CreateType uint8

type Template struct {
	Id         int64                    `json:"id"`
	Name       string                   `json:"name"`
	CreateType CreateType               `json:"create_type"`
	Rules      []map[string]interface{} `json:"rules"`
	Options    map[string]interface{}   `json:"options"`
	Desc       string                   `json:"desc"`
}

type RetrieveTemplates struct {
	Total     int64      `json:"total"`
	Templates []Template `json:"templates"`
}
