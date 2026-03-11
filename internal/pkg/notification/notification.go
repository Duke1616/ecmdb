package notification

import (
	"github.com/Duke1616/ecmdb/internal/workflow"
)

type NotificationResponse struct {
	// 通知平台生成的通知ID
	NotificationId int64 `json:"notification_Id"`
	// 发送状态
	Status string `json:"status"`
	// 失败时的错误代码
	ErrorCode string `json:"error_code"`
	// 错误详情
	ErrorMsg string `json:"error_msg"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(notificationId int64, status string) NotificationResponse {
	return NotificationResponse{
		NotificationId: notificationId,
		Status:         status,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(errorCode, errorMsg string) NotificationResponse {
	return NotificationResponse{
		ErrorCode: errorCode,
		ErrorMsg:  errorMsg,
	}
}

// NewErrorResponseWithID 创建带通知ID的错误响应
func NewErrorResponseWithID(notificationId int64, status, errorCode, errorMsg string) NotificationResponse {
	return NotificationResponse{
		NotificationId: notificationId,
		Status:         status,
		ErrorCode:      errorCode,
		ErrorMsg:       errorMsg,
	}
}

type Notification struct {
	Receiver     string   `json:"receiver"`      // 接收者(手机/邮箱/用户ID)
	ReceiverType string   `json:"receiver_type"` // 接收者类型 (user_id, open_id, chat_id等)
	Template     Template `json:"template"`      // 发送模版
	Channel      Channel  `json:"channel"`       // 发送渠道
	WorkFlowID   int64    `json:"workflow_id"`   // 工作流定义ID
	MessageID    string   `json:"message_id"`    // 消息ID（用于更新消息）
}

const (
	ReceiverTypeUser      = "user_id"
	ReceiverTypeChatGroup = "chat_id"
)

func (n Notification) IsPatch() bool {
	return n.MessageID != ""
}

func (n Notification) IsProgressImageResult() bool {
	return n.Template.Name == workflow.NotifyTypeProgressImageResult
}

type Channel string

const (
	ChannelLarkCard Channel = "LARK_CARD"
	ChannelLarkText Channel = "LARK_TEXT"
	ChannelWechat   Channel = "WECHAT"
)

func (c Channel) String() string {
	return string(c)
}

type FieldType string

type InputOption struct {
	Label string `json:"label"` // 选项显示名
	Value string `json:"value"` // 选项值
}

type InputField struct {
	Name     string            `json:"name"`     // 表单字段显示名
	Key      string            `json:"key"`      // 表单字段键名（对应 Order Data Key）
	Type     FieldType         `json:"type"`     // 字段类型：input, textarea, date, number...
	Required bool              `json:"required"` // 是否必填
	Options  []InputOption     `json:"options"`  // 选项列表（用于 select 等）
	Props    map[string]string `json:"props"`    // 额外组件属性（如 placeholder）
	Value    string            `json:"value"`    // 数据值
	ReadOnly bool              `json:"readonly"` // 只读字段，比如提示用户时候使用
}

type Template struct {
	Name        workflow.NotifyType `json:"name"`         // 模版名称
	Title       string              `json:"title"`        // 模版标题
	Fields      []Field             `json:"fields"`       // 模版字段信息
	Values      []Value             `json:"values"`       // 模版传递变量
	InputFields []InputField        `json:"input_fields"` // 录入的字段
	HideForm    bool                `json:"hide_form"`    // 隐藏
	Remark      string              `json:"remark"`       // 备注信息
	ImageKey    string              `json:"image_key"`    // 图片地址
	Text        string              `json:"text"`         // 文本信息
}

type Filter struct {
	UserIds     []string `json:"user_ids"`
	ProjectIds  []string `json:"project_ids"`
	ResourceIds []string `json:"resource_ids"`
}

// AddRowSpacers 为传入的数据展示卡片阵列增加双列排版时的视觉行间距。
// 每呈现两项（占满飞书一横排）后，追加一个全宽的空文本节点进行物理换行。
func AddRowSpacers(fields []Field) []Field {
	var results []Field
	for i, f := range fields {
		results = append(results, f)
		if (i+1)%2 == 0 {
			results = append(results, Field{
				IsShort: false,
				Tag:     "lark_md",
				Content: "",
			})
		}
	}
	return results
}

type Field struct {
	IsShort   bool   `json:"is_short"`
	IsDivider bool   `json:"is_divider"`
	Tag       string `json:"tag"`
	Content   string `json:"content"`
}

type Value struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}
