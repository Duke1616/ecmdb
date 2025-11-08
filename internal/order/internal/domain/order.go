package domain

import (
	"fmt"

	"github.com/Duke1616/ecmdb/internal/order/internal/errs"
)

type Channel string

func (c Channel) String() string {
	return string(c)
}

const (
	ChannelFeishuCard Channel = "FEISHU_CARD" // 短信
	ChannelEmail      Channel = "EMAIL"       // 邮件
	ChannelInApp      Channel = "IN_APP"      // 站内信
)

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
	// ALERT 告警信息
	ALERT Provide = 3
)

func (s Provide) IsValid() bool {
	return s == SYSTEM || s == WECHAT || s == ALERT
}

func (s Provide) IsAlert() bool {
	return s == ALERT
}

type Order struct {
	Id               int64
	BizID            int64  // 业务ID
	Key              string // 业务唯一 Key
	TemplateId       int64
	WorkflowId       int64
	Data             map[string]interface{}
	Status           Status
	Provide          Provide
	CreateBy         string
	Process          Process
	Ctime            int64
	Wtime            int64
	NotificationConf NotificationConf // 为了引入告警转工单，引入外部消息通知
}

// NotificationConf 消息通知配置
type NotificationConf struct {
	TemplateID     int64                  // 模版ID
	TemplateParams map[string]interface{} // 传递参数
	Channel        Channel                // 通知渠道
}

func (o *Order) Validate() error {
	if o.TemplateId <= 0 {
		return fmt.Errorf("%w: Template.ID = %d", errs.ErrInvalidParameter, o.TemplateId)
	}

	if o.WorkflowId <= 0 {
		return fmt.Errorf("%w: WorkFlow.ID = %d", errs.ErrInvalidParameter, o.WorkflowId)
	}

	if !o.Provide.IsValid() {
		return fmt.Errorf("%w: 不支持的来源提供商", errs.ErrInvalidParameter)
	}

	if o.CreateBy == "" {
		return fmt.Errorf("%w: 工单创建人不能为空", errs.ErrInvalidParameter)
	}

	return nil
}

type Process struct {
	InstanceId int
}
