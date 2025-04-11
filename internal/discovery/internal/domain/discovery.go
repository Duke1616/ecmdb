package domain

// Discovery 模版自动化节点、自动发现调度节点
type Discovery struct {
	Id         int64
	TemplateId int64
	RunnerId   int64
	RunnerName string
	Field      string
	Title      string
	Value      string
}
