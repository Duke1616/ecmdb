package web

type CreateDiscoveryReq struct {
	TemplateId int64  `json:"template_id"`
	RunnerId   int64  `json:"runner_id"`
	RunnerName string `json:"runner_name"` // 前端展示
	Field      string `json:"field"`
	Title      string `json:"title"` // 前端展示
	Value      string `json:"value"`
}

type UpdateDiscoveryReq struct {
	Id         int64  `json:"id"`
	RunnerId   int64  `json:"runner_id"`
	RunnerName string `json:"runner_name"`
	Field      string `json:"field"`
	Title      string `json:"title"`
	Value      string `json:"value"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListByTemplateId struct {
	Page
	TemplateId int64 `json:"template_id"`
}

type Discovery struct {
	Id         int64  `json:"id"`
	TemplateId int64  `json:"template_id"`
	RunnerId   int64  `json:"runner_id"`
	RunnerName string `json:"runner_name"`
	Field      string `json:"field"`
	Title      string `json:"title"`
	Value      string `json:"value"`
}

type RetrieveDiscoveries struct {
	Total       int64       `json:"total"`
	Discoveries []Discovery `json:"discoveries"`
}

type DeleteDiscoveryReq struct {
	Id int64 `json:"id"`
}
