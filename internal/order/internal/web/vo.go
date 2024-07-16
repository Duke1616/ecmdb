package web

import "github.com/Bunny3th/easy-workflow/workflow/database"

type CreateOrderReq struct {
	CreateBy   string                 `json:"create_by"`
	TemplateId int64                  `json:"template_id"`
	WorkflowId int64                  `json:"workflow_id"`
	Data       map[string]interface{} `json:"data"`
}

type DetailReq struct {
	ProcessInstanceId int `json:"process_instance_id"`
}

type Todo struct {
	UserId      string `json:"user_id"`
	ProcessName string `json:"proc_name"`
	SortByAsc   bool   `json:"sort_by_asc" validate:"required"`
	Offset      int    `json:"offset,omitempty"`
	Limit       int    `json:"limit,omitempty" validate:"required"`
}

type StartUser struct {
	UserId      string `json:"user_id"`
	ProcessName string `json:"proc_name"`
	Offset      int    `json:"offset,omitempty"`
	Limit       int    `json:"limit,omitempty" validate:"required"`
}

type MyOrderReq struct {
	CreateBy string `json:"create_by"`
}

type Order struct {
	TaskId             int                 `json:"task_id"`               // 任务ID
	ProcessInstanceId  int                 `json:"process_instance_id"`   // 流程实例ID
	Starter            string              `json:"starter"`               // 提单人
	Title              string              `json:"title"`                 // 标题
	CurrentStep        string              `json:"current_step"`          // 当前步骤
	ApprovedBy         []string            `json:"approved_by"`           // 当前处理人
	ProcInstCreateTime *database.LocalTime `json:"proc_inst_create_time"` // 流程开始时间
	Ctime              int64               `json:"ctime"`                 // 创建工单时间
	TemplateId         int64               `json:"template_id"`
	WorkflowId         int64               `json:"workflow_id"`
}

type RetrieveOrders struct {
	Total int64   `json:"total"`
	Tasks []Order `json:"orders"`
}
