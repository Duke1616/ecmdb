package event

const TaskRunnerEventName = "task_runner_events"

type Action uint8

type TaskRunnerEvent struct {
	TaskIdentify string
	TaskSecret   string
	WorkName     string
	Name         string
	Tags         []string
	Desc         string
	Action       Action
}
