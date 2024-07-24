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
	// RUNNING 运行中
	RUNNING Status = 3
	// WAITING 等待运行
	WAITING Status = 3
)

type Task struct {
	Id            int64
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
