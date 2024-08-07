package domain

type Status uint8

func (s Status) ToUint8() uint8 {
	return uint8(s)
}

const (
	// START 等待开始
	START Status = 1
	// PROCESS 流程运行中
	PROCESS Status = 2
	// END 完成
	END Status = 3
	// RETRY 重试
	RETRY Status = 4
	// AUTOMATION 自动化
	AUTOMATION Status = 5
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
}

type Process struct {
	InstanceId int
}
