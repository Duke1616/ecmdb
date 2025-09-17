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
	Name           string      `json:"name"`
	CodebookUid    string      `json:"codebook_uid"`
	CodebookSecret string      `json:"codebook_secret"`
	WorkerName     string      `json:"worker_name"`
	Topic          string      `json:"topic"`
	Tags           []string    `json:"tags"`
	Desc           string      `json:"desc"`
	Variables      []Variables `json:"variables"`
}

type UpdateRunnerReq struct {
	Id             int64       `json:"id"`
	Name           string      `json:"name"`
	CodebookUid    string      `json:"codebook_uid"`
	CodebookSecret string      `json:"codebook_secret"`
	WorkerName     string      `json:"worker_name"`
	Topic          string      `json:"topic"`
	Tags           []string    `json:"tags"`
	Desc           string      `json:"desc"`
	Variables      []Variables `json:"variables"`
}

type ListByWorkflowIdReq struct {
	WorkflowId int64 `json:"workflow_id"`
}

type DeleteRunnerReq struct {
	Id int64 `json:"id"`
}

type ListRunnerByIds struct {
	Ids []int64 `json:"ids"`
}
type Variables struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Secret bool   `json:"secret"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListRunnerReq struct {
	Page
}

type Runner struct {
	Id          int64       `json:"id"`
	Name        string      `json:"name"`
	CodebookUid string      `json:"codebook_uid"`
	WorkerName  string      `json:"worker_name"`
	Tags        []string    `json:"tags"`
	Variables   []Variables `json:"variables"`
	Desc        string      `json:"desc"`
}

type RetrieveWorkers struct {
	Total   int64    `json:"total"`
	Runners []Runner `json:"runners"`
}

type RunnerTags struct {
	CodebookName     string            `json:"codebook_name"`
	CodebookUid      string            `json:"codebook_uid"`
	TagsMappingTopic map[string]string `json:"tags_topic"`
}

type RetrieveRunnerTags struct {
	RunnerTags []RunnerTags `json:"runner_tags"`
}
