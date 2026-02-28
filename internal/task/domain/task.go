package domain

import "time"

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

type Status uint8

func (s Status) ToUint8() uint8 {
	return uint8(s)
}

const (
	// SUCCESS 成功
	SUCCESS Status = 1
	// FAILED 失败
	FAILED Status = 2
	// RUNNING 运行中
	RUNNING Status = 3
	// WAITING 等待/初始化: 流程刚流转到该节点，刚入库
	WAITING Status = 4
	// BLOCKED 挂起/阻塞: 找不到执行路由或参数缺失等异常，流程无法推进
	BLOCKED Status = 5
	// SCHEDULED 已分配/已就绪: 任务已提交给 Dispatcher，或处于定时等待触发的状态
	SCHEDULED Status = 6
)

type Task struct {
	Id            int64
	ProcessInstId int
	// 触发位置、比如错误等
	TriggerPosition string
	CurrentNodeId   string
	OrderId         int64
	CodebookName    string
	CodebookUid     string
	WorkflowId      int64
	Code            string
	Language        string
	Result          string
	WantResult      string
	Status          Status
	IsTiming        bool
	Utime           int64
	ScheduledTime   int64
	Variables       []Variables
	Args            map[string]interface{}
	Kind            Kind
	Target          string // 执行目标 (Topic 或 ServiceName)
	Handler         string // 执行方法
	ExternalId      string // 外部任务 ID (如分布式任务平台生成的实例 ID)
	StartTime       int64  // 任务实际开始执行时间
	EndTime         int64  // 任务完成或失败时间
	RetryCount      int    // 自动重试次数，超过阈值后转为 BLOCKED 等待人工干预
}

type TaskResult struct {
	Id              int64  `json:"id"`
	TriggerPosition string `json:"trigger_position"`
	WantResult      string `json:"want_result"`
	Result          string `json:"result"`
	Status          Status `json:"status"`
	time            time.Time
	StartTime       int64
	EndTime         int64
	RetryCount      int // 每次重试时传 1，由 DAO 用 $inc 原子递增
}

type Variables struct {
	Key    string
	Value  string
	Secret bool
}
