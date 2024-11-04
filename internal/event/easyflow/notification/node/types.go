package node

import (
	"github.com/Duke1616/ecmdb/internal/engine"
	"github.com/Duke1616/ecmdb/internal/event/easyflow/notification/method"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/task"
	"github.com/Duke1616/ecmdb/internal/template"
	"github.com/Duke1616/ecmdb/internal/user"
)

type MyNotifications struct {
	engineSvc   engine.Service
	templateSvc template.Service
	orderSvc    order.Service
	userSvc     user.Service
	taskSvc     task.Service
	integration []method.NotifyIntegration
}

//func (n *MyNotifications) GetAutomation() notification.User {
//	// 返回一个实现了 User 接口的实例
//	return &MyUser{}
//}
//
//func (n *MyNotifications) GetUser() notification.Automation {
//	// 返回一个实现了 Automation 接口的实例
//	return &MyAutomation{}
//}
