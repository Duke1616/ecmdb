package ioc

import (
	"github.com/Duke1616/ecmdb/internal/event/resource"
)

// InitTasks 初始化所有后台任务
// NOTE: 新增后台任务时在此处注入，打通定时任务、后台作业补偿及全量大事件 Kafka 消费监听
func InitTasks(
	fieldDeleteConsumer *resource.FieldDeleteConsumer,
	fieldSecretConsumer *resource.FieldSecureAttrChangeConsumer,

) []Task {
	return []Task{
		fieldDeleteConsumer,
		fieldSecretConsumer,
	}
}
