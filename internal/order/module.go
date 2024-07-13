package order

import (
	"github.com/Duke1616/ecmdb/internal/order/internal/event/consumer"
	"github.com/Duke1616/ecmdb/internal/order/internal/web"
)

type Module struct {
	Hdl *web.Handler
	Svc Service
	c   *consumer.WechatOrderConsumer
}
