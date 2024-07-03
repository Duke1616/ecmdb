package domain

type Workflow struct {
	Id         int64
	TemplateId int64
	Name       string
	Icon       string
	Owner      string
	Desc       string
	FlowData   map[string]interface{}
}
