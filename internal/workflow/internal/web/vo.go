package web

type CreateReq struct {
	TemplateId   int64     `json:"template_id"`
	Name         string    `json:"name"`
	Icon         string    `json:"icon"`
	Owner        string    `json:"owner"`
	Desc         string    `json:"desc"`
	IsNotify     bool      `json:"is_notify"`
	NotifyMethod uint8     `json:"notify_method"`
	FlowData     LogicFlow `json:"flow_data"`
}

type LogicFlow struct {
	Edges []map[string]interface{} `json:"edges"`
	Nodes []map[string]interface{} `json:"nodes"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListReq struct {
	Page
}

type DeployReq struct {
	Id int64
}

type UpdateReq struct {
	Id           int64     `json:"id"`
	Name         string    `json:"name"`
	Desc         string    `json:"desc"`
	Owner        string    `json:"owner"`
	IsNotify     bool      `json:"is_notify"`
	NotifyMethod uint8     `json:"notify_method"`
	FlowData     LogicFlow `json:"flow_data"`
}

type DeleteReq struct {
	Id int64 `json:"id"`
}

type Workflow struct {
	Id           int64     `json:"id"`
	TemplateId   int64     `json:"template_id"`
	Name         string    `json:"name"`
	Icon         string    `json:"icon"`
	Owner        string    `json:"owner"`
	Desc         string    `json:"desc"`
	IsNotify     bool      `json:"is_notify"`
	NotifyMethod uint8     `json:"notify_method"`
	FlowData     LogicFlow `json:"flow_data"`
}

type RetrieveWorkflows struct {
	Total     int64      `json:"total"`
	Workflows []Workflow `json:"workflows"`
}

type OrderGraphReq struct {
	Id                int64 `json:"id"`
	Status            uint8 `json:"status"`
	ProcessInstanceId int   `json:"process_instance_id"`
}

type RetrieveOrderGraph struct {
	EdgeIds  []string `json:"edge_ids"`
	Workflow Workflow `json:"workflow"`
}
