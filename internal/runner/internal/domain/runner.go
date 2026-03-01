package domain

type Kind string

const (
	// KAFKA 工作节点 kafka 推送
	KAFKA Kind = "KAFKA"
	// GRPC 绑定分布式任务平台执行节点
	GRPC Kind = "GRPC"
)

func (s Kind) ToString() string {
	return string(s)
}

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

type Runner struct {
	Id             int64
	Name           string
	CodebookUid    string
	CodebookSecret string
	Tags           []string // 绑定标签，自动化任务通过标签进行匹配
	Desc           string
	Kind           Kind   // 运行模式
	Target         string // 执行目标 (Topic 或 ServiceName)
	Handler        string // 执行方法
	Action         Action
	Variables      []Variables
}

func (r Runner) IsKindKafka() bool {
	return r.Kind == KAFKA
}

type Variables struct {
	Key    string
	Value  string
	Secret bool
}

type TagDetail struct {
	Kind    Kind
	Target  string
	Handler string
}

type RunnerTags struct {
	CodebookUid string
	// Tag -> Detail info
	TagsMapping map[string]TagDetail
}
