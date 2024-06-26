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

type RunnerEvent struct {
	Name     string // 执行名称
	UUID     string // 唯一标识
	Language string // 语言
	Code     string // 代码
}
