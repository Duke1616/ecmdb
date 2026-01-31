package order

import (
	"github.com/Duke1616/ecmdb/internal/order/internal/event/consumer"
	"github.com/Duke1616/ecmdb/internal/order/internal/web"
)

type Module struct {
	Hdl       *web.Handler
	RpcServer *RpcServer
	Svc       Service
	EngineSvc EngineService
	cw        *consumer.WechatOrderConsumer
	cs        *consumer.ProcessEventConsumer
	cms       *consumer.OrderStatusModifyEventConsumer
	cf        *consumer.LarkCallbackEventConsumer
}
