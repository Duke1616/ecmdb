package domain

type Status uint8

func (s Status) ToUint8() uint8 {
	return uint8(s)
}

func (s Status) ToInt() int {
	return int(s)
}

const (
	// START 等待开始
	START Status = 1
	// PROCESS 流程运行中
	PROCESS Status = 2
	// END 完成
	END Status = 3
	// WITHDRAW 撤回
	WITHDRAW Status = 4
)

type Provide uint8

func (s Provide) ToUint8() uint8 {
	return uint8(s)
}

const (
	// SYSTEM 本系统
	SYSTEM Provide = 1
	// WECHAT 企业微信
	WECHAT Provide = 2
)

type Order struct {
	Id           int64
	TemplateId   int64
	TemplateName string
	WorkflowId   int64
	Data         map[string]interface{}
	Status       Status
	Provide      Provide
	CreateBy     string
	Process      Process
	Ctime        int64
	Wtime        int64
}

type Process struct {
	InstanceId int
}
