package event

import (
	"context"

	"github.com/Duke1616/ecmdb/pkg/mqx"
	"github.com/ecodeclub/mq-api"
)

// FIELD_DELETE_EVENT_NAME 字段删除事件的 Topic 名称
const FIELD_DELETE_EVENT_NAME = "field_delete_event"

// FieldDelete 字段删除事件数据结构
type FieldDelete struct {
	ModelUid    string `json:"model_uid"`    // 模型唯一标识
	FieldUid    string `json:"field_uid"`    // 字段唯一标识
	TriggerTime int64  `json:"trigger_time"` // 触发时间（毫秒戳）
}

// IFieldDeleteEventProducer 字段删除事件发布者接口
type IFieldDeleteEventProducer interface {
	// Produce 发布字段删除事件
	Produce(ctx context.Context, evt FieldDelete) error
}

type fieldDeleteEventProducer struct {
	producer mq.Producer
}

// NewFieldDeleteEventProducer 构造字段删除事件发布者
func NewFieldDeleteEventProducer(q mq.MQ) (IFieldDeleteEventProducer, error) {
	// NOTE: 使用 mqx 提供的泛型通用发布器构造
	return mqx.NewGeneralProducer[FieldDelete](q, FIELD_DELETE_EVENT_NAME)
}
