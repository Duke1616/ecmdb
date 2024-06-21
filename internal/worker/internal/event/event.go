package event

const TaskWorkerEventName = "task_worker_events"

type WorkerEvent struct {
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Topic string `json:"topic"`
}
