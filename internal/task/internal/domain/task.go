package domain

import "time"

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
	WAITING Status = 4
	// PENDING 运行某处意外无法执行
	PENDING Status = 5
	// SCHEDULE 等待调度
	SCHEDULE Status = 6
	//RETRY 重试
	RETRY Status = 7
	// TIMING 定时任务
	TIMING Status = 8
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
	WorkerName      string
	WorkflowId      int64
	Code            string
	Topic           string
	Language        string
	Result          string
	WantResult      string
	Status          Status
	IsTiming        bool
	Utime           int64
	Timing          Timing
	Variables       []Variables
	Args            map[string]interface{}
}

type Unit uint8

func (s Unit) ToUint8() uint8 {
	return uint8(s)
}

const (
	// MINUTE 分钟
	MINUTE Unit = 1
	// HOUR 小时
	HOUR Unit = 2
	// DAY 天
	DAY Unit = 3
)

// Timing 定时执行
type Timing struct {
	Stime    int64
	Unit     Unit
	Quantity int64
}

type TaskResult struct {
	Id              int64  `json:"id"`
	TriggerPosition string `json:"trigger_position"`
	WantResult      string `json:"want_result"`
	Result          string `json:"result"`
	Status          Status `json:"status"`
	time            time.Time
}

type Variables struct {
	Key    string
	Value  any
	Secret bool
}
