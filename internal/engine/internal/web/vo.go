package web

import (
	"github.com/Bunny3th/easy-workflow/workflow/database"
)

type StartTaskReq struct {
	ProcessId  int         `json:"process_id"`
	BusinessId string      `json:"business_id"`
	Comment    string      `json:"comment"`
	Variables  []Variables `json:"variables"`
}

type Variables struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TodoListTaskReq struct {
	UserId      string `json:"user_id"`
	ProcessName string `json:"proc_name"`
	SortByAsc   bool   `json:"sort_by_asc" validate:"required"`
	Idx         int    `json:"idx"`
	Rows        int    `json:"rows" validate:"required"`
}

type PassTaskReq struct {
	TaskId        int    `json:"task_id"`
	Comment       string `json:"comment"`
	VariablesJson string `json:"variables_json"`
}

type Task struct {
	TaskId             int                 `json:"task_id"`               // 任务ID
	ProcessInstanceId  int                 `json:"process_instance_id"`   // 流程创建ID
	Starter            string              `json:"starter"`               // 提单人
	Title              string              `json:"title"`                 // 标题
	CurrentStep        string              `json:"current_step"`          // 当前步骤
	ApprovedBy         []string            `json:"approved_by"`           // 当前处理人
	ProcInstCreateTime *database.LocalTime `json:"proc_inst_create_time"` // 流程开始时间
}

type RetrieveTasks struct {
	Total int64  `json:"total"`
	Tasks []Task `json:"tasks"`
}
