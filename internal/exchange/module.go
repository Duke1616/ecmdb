package exchange

import (
	"github.com/Duke1616/ecmdb/internal/exchange/internal/service"
	"github.com/Duke1616/ecmdb/internal/exchange/internal/web"
)

// IExchangeService 数据交换服务接口
type IExchangeService = service.IExchangeService

// Handler Web层处理器
type Handler = web.Handler

type Module struct {
	Hdl *Handler
}
