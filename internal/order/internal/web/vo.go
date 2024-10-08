package web

import (
	"github.com/Bunny3th/easy-workflow/workflow/database"
)

type CreateOrderReq struct {
	CreateBy     string                 `json:"create_by"`
	TemplateId   int64                  `json:"template_id"`
	TemplateName string                 `json:"template_name"`
	WorkflowId   int64                  `json:"workflow_id"`
	Data         map[string]interface{} `json:"data"`
}

type DetailProcessInstIdReq struct {
	ProcessInstanceId int `json:"process_instance_id"`
}

type Todo struct {
	UserId      string `json:"user_id"`
	ProcessName string `json:"proc_name"`
	SortByAsc   bool   `json:"sort_by_asc" validate:"required"`
	Offset      int    `json:"offset,omitempty"`
	Limit       int    `json:"limit,omitempty" validate:"required"`
}

type HistoryReq struct {
	UserId      string `json:"user_id"`
	ProcessName string `json:"proc_name"`
	SortByAsc   bool   `json:"sort_by_asc" validate:"required"`
	Offset      int64  `json:"offset,omitempty"`
	Limit       int64  `json:"limit,omitempty" validate:"required"`
}

type PassOrderReq struct {
	TaskId  int    `json:"task_id"`
	Comment string `json:"comment"`
}

type RejectOrderReq struct {
	TaskId  int    `json:"task_id"`
	Comment string `json:"comment"`
}

type RecordTaskReq struct {
	ProcessInstId int `json:"process_inst_id"`
	Offset        int `json:"offset,omitempty"`
	Limit         int `json:"limit,omitempty" validate:"required"`
}

type StartUser struct {
	UserId      string `json:"user_id"`
	ProcessName string `json:"proc_name"`
	Offset      int    `json:"offset,omitempty"`
	Limit       int    `json:"limit,omitempty" validate:"required"`
}

type StartUserReq struct {
	ProcessInstId int    `json:"process_inst_id"`
	Starter       string `json:"starter"`
	Offset        int64  `json:"offset,omitempty"`
	Limit         int64  `json:"limit,omitempty"`
}

type RevokeOrderReq struct {
	InstanceId int  `json:"instance_id"`
	Force      bool `json:"force"`
}

type MyOrderReq struct {
	CreateBy string `json:"create_by"`
}

type Order struct {
	Id                 int64                  `json:"id"`
	TaskId             int                    `json:"task_id"`             // 任务ID
	ProcessInstanceId  int                    `json:"process_instance_id"` // 流程实例ID
	Starter            string                 `json:"starter"`             // 提单人
	TemplateName       string                 `json:"template_name"`       // 模版名称
	Provide            uint8                  `json:"provide"`
	CurrentStep        string                 `json:"current_step"`
	ApprovedBy         string                 `json:"approved_by"`           // 处理人
	ProcInstCreateTime *database.LocalTime    `json:"proc_inst_create_time"` // 流程开始时间
	Ctime              string                 `json:"ctime"`                 // 创建工单时间
	Wtime              string                 `json:"wtime"`                 // 工单完成时间
	TemplateId         int64                  `json:"template_id"`
	WorkflowId         int64                  `json:"workflow_id"`
	Data               map[string]interface{} `json:"data"`
}

type Steps struct {
	CurrentStep string   `json:"current_step"`
	ApprovedBy  []string `json:"approved_by"` // 处理人
}

type RetrieveOrders struct {
	Total int64   `json:"total"`
	Tasks []Order `json:"orders"`
}

type TaskRecord struct {
	Nodename     string              `json:"nodename"`      // 当前步骤
	ApprovedBy   string              `json:"approved_by"`   // 处理人
	IsCosigned   int                 `json:"is_cosigned"`   // 是否会签
	Status       int                 `json:"status"`        // 任务状态:0:初始 1:通过 2:驳回
	Comment      string              `json:"comment"`       // 评论
	IsFinished   int                 `json:"is_finished"`   // 0:任务未完成 1:处理完成
	FinishedTime *database.LocalTime `json:"finished_time"` // 处理任务时间
}

type RetrieveTaskRecords struct {
	TaskRecords []TaskRecord `json:"task_records"`
	Total       int64        `json:"total"`
}
