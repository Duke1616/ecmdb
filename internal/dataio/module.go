package dataio

import (
	"github.com/Duke1616/ecmdb/internal/dataio/internal/service"
	"github.com/Duke1616/ecmdb/internal/dataio/internal/web"
)

// IDataIOService 数据导入导出服务接口
type IDataIOService = service.IDataIOService

// Handler Web层处理器
type Handler = web.Handler

type Module struct {
	Hdl *Handler
}
