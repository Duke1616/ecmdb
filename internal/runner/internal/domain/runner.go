package domain

type RunMode string

const (
	// RunModeWorker 工作节点 kafka 推送
	RunModeWorker RunMode = "WORKER"
	// RunModeExecute 绑定分布式任务平台执行节点
	RunModeExecute RunMode = "EXECUTE"
)

func (s RunMode) ToString() string {
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
	RunMode        RunMode // 运行模式
	Worker         *Worker
	Execute        *Execute
	Action         Action
	Variables      []Variables
}

func (r Runner) IsModeWorker() bool {
	return r.RunMode == RunModeWorker
}

type Worker struct {
	WorkerName string // 工作节点名出
	Topic      string // kafka topic 队列
}

type Execute struct {
	ServiceName string // 执行器名称
	Handler     string // 执行器方法
}

type Variables struct {
	Key    string
	Value  string
	Secret bool
}

type RunnerTags struct {
	CodebookUid      string
	TagsMappingTopic map[string]string
}
