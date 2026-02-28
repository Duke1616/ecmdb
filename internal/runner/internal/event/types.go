package event

const TaskRegisterRunnerEventName = "register_runner_event"

type Action uint8

type TaskRunnerEvent struct {
	CodebookUid    string
	CodebookSecret string
	WorkerName     string
	Topic          string
	Handler        string
	Name           string
	Tags           []string
	Desc           string
	Action         Action
}
