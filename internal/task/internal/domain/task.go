package domain

type Status uint8

func (s Status) ToUint8() uint8 {
	return uint8(s)
}

const (
	// SUCCESS 成功
	SUCCESS Status = 1
	// FAILED 失败
	failed
	FAILED Status = 2
)

type Task struct {
	ProcessInstId int
	OrderId       int64
	CodebookUid   string
	WorkerName    string
	WorkflowId    int64
	Code          string
	Topic         string
	Language      string
	Result        string
	Status        Status
}

type TaskResult struct {
	Id     int64  `json:"id"`
	Result string `json:"result"`
	Status Status `json:"status"`
}
