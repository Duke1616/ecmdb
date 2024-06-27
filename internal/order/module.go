package order

import (
	"github.com/Duke1616/ecmdb/internal/order/internal/event"
	"github.com/Duke1616/ecmdb/internal/order/internal/web"
)

type Module struct {
	Hdl *web.Handler
	Svc Service
	c   *event.WechatOrderConsumer
}
