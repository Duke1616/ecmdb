package workflow

import (
	"github.com/Duke1616/ecmdb/internal/workflow/internal/domain"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/service"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Workflow = domain.Workflow

type NotifyMethod = domain.NotifyMethod

// NotifyMethodToString 将 NotifyMethod 转换为对应的文字描述
func NotifyMethodToString(method NotifyMethod) string {
	switch method {
	case domain.Feishu:
		return "feishu"
	case domain.Wechat:
		return "wechat"
	default:
		return "Unknown"
	}
}
