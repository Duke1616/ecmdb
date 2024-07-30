package event

const TaskRegisterRunnerEventName = "register_runner_event"

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
