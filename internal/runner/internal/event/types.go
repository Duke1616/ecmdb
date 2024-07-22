package event

const TaskRunnerEventName = "task_runner_events"

type Action uint8

type TaskRunnerEvent struct {
	CodebookUid    string
	CodebookSecret string
	WorkerName     string
	Name           string
	Tags           []string
	Desc           string
	Action         Action
}
