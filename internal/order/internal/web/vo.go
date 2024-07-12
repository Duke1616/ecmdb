package web

type CreateOrderReq struct {
	CreateBy   string                 `json:"create_by"`
	TemplateId int64                  `json:"template_id"`
	FlowId     int64                  `json:"flow_id"`
	Data       map[string]interface{} `json:"data"`
}
