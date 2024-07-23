package domain

type Task struct {
	ProcessInstId int
	OrderId       int64
	CodebookUid   string
	WorkerName    string
	WorkflowId    int64
	Code          string
	Topic         string
	Language      string
}
