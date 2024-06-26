package web

type Action uint8

func (s Action) ToUint8() uint8 {
	return uint8(s)
}

const (
	// REGISTER 注册
	REGISTER Action = 1
	// UNREGISTER 注销
	UNREGISTER Action = 2
)

type RegisterRunnerReq struct {
	Name           string   `json:"name"`
	TaskIdentifier string   `json:"task_identifier"`
	TaskSecret     string   `json:"task_secret"`
	WorkName       string   `json:"work_name"`
	Tags           []string `json:"tags"`
	Desc           string   `json:"desc"`
	Action         Action   `json:"action"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListRunnerReq struct {
	Page
}

type Runner struct {
	Id   int64    `json:"id"`
	Name string   `json:"name"`
	Tags []string `json:"tags"`
	Desc string   `json:"desc"`
}

type RetrieveWorkers struct {
	Total   int64    `json:"total"`
	Runners []Runner `json:"runners"`
}
