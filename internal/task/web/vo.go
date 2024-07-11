package web

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
	Idx         int    `json:"idx" validate:"required"`
	Rows        int    `json:"rows" validate:"required"`
}

type PassTaskReq struct {
	TaskId        int    `json:"task_id"`
	Comment       string `json:"comment"`
	VariablesJson string `json:"variables_json"`
}

type Task struct {
	NodeName string `json:"nodename"`
}

type RetrieveTasks struct {
}
