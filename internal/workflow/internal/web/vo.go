package web

type CreateReq struct {
	TemplateId int64                  `json:"template_id"`
	Name       string                 `json:"name"`
	Icon       string                 `json:"icon"`
	Owner      string                 `json:"owner"`
	Desc       string                 `json:"desc"`
	FlowData   map[string]interface{} `json:"flow_data"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListReq struct {
	Page
}

type Workflow struct {
	Id         int64                  `json:"id"`
	TemplateId int64                  `json:"template_id"`
	Name       string                 `json:"name"`
	Icon       string                 `json:"icon"`
	Owner      string                 `json:"owner"`
	Desc       string                 `json:"desc"`
	FlowData   map[string]interface{} `json:"flow_data"`
}

type RetrieveWorkflows struct {
	Total     int64      `json:"total"`
	Workflows []Workflow `json:"workflows"`
}
