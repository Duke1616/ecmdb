package event

const TaskWorkerEventName = "task_worker_events"

type Status uint8

func (s Status) ToUint8() uint8 {
	return uint8(s)
}

const (
	// RUNNING 启用
	RUNNING Status = 1
	// STOPPING 停止
	STOPPING Status = 2
)

type WorkerEvent struct {
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Topic  string `json:"topic"`
	Status Status `json:"status"`
}

type AgentExecuteEvent struct {
	TaskId    int64                  `json:"task_id"`
	Handler   string                 `json:"handler"`
	Language  string                 `json:"language"`
	Code      string                 `json:"code"`
	Args      map[string]interface{} `json:"args"`
	Variables string                 `json:"variables"`
}
