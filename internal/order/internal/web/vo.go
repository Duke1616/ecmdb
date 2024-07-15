package web

type CreateOrderReq struct {
	CreateBy   string                 `json:"create_by"`
	TemplateId int64                  `json:"template_id"`
	WorkflowId int64                  `json:"workflow_id"`
	Data       map[string]interface{} `json:"data"`
}

type Todo struct {
	UserId      string `json:"user_id"`
	ProcessName string `json:"proc_name"`
	SortByAsc   bool   `json:"sort_by_asc" validate:"required"`
	Idx         int    `json:"idx"`
	Rows        int    `json:"rows" validate:"required"`
}
