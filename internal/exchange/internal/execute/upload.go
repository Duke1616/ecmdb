package execute

import (
	"github.com/Duke1616/ecmdb/internal/exchange/internal/service"
	"github.com/Duke1616/ework-runner/sdk/executor"
)

type parserHandler struct {
	svc service.IExchangeService
}

func (h *parserHandler) Name() string {
	return "parser-excel"
}

func (h *parserHandler) Run(ctx *executor.Context) error {
	// 1. 根据传递的 excel 信息下载文件

	// 2. 根据文件的内容，动态录入到 CMDB 平台中

	return nil
}
