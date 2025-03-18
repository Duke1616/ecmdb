package web

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

type RegisterRunnerReq struct {
	Name           string   `json:"name"`
	CodebookUid    string   `json:"codebook_uid"`
	CodebookSecret string   `json:"codebook_secret"`
	WorkerName     string   `json:"worker_name"`
	Tags           []string `json:"tags"`
	Desc           string   `json:"desc"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListTaskReq struct {
	Page
}

type ListTaskByInstanceIdReq struct {
	InstanceId int `json:"instance_id"`
	Page
}

type Task struct {
	Id           int64  `json:"id"`
	OrderId      int64  `json:"order_id"`
	WorkerName   string `json:"worker_name"`
	CodebookUid  string `json:"codebook_uid"`
	CodebookName string `json:"codebook_name"`
	Status       Status `json:"status"`
	IsTiming     bool   `json:"is_timing"`
	Timing       Timing `json:"timing"`
	StartTime    string `json:"start_time"`
	Code         string `json:"code"`
	Language     string `json:"language"`
	Args         string `json:"args"`
	Variables    string `json:"variables"`
	Result       string `json:"result"`
}

type Timing struct {
	Stime    int64 `json:"stime"`
	Unit     uint8 `json:"unit"`
	Quantity int64 `json:"quantity"`
}

type RetryReq struct {
	Id int64 `json:"id"`
}

type UpdateArgsReq struct {
	Id   int64                  `json:"id"`
	Args map[string]interface{} `json:"args"`
}

type UpdateVariablesReq struct {
	Id        int64  `json:"id"`
	Variables string `json:"variables"`
}

type RetrieveTasks struct {
	Total int64  `json:"total"`
	Tasks []Task `json:"tasks"`
}

type UpdateStatusToSuccessReq struct {
	Id int64 `json:"id"`
}

type Variables struct {
	Key    string `json:"key"`
	Value  any    `json:"value"`
	Secret bool   `json:"secret"`
}
