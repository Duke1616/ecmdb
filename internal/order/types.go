package order

import (
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/order/internal/web"
)

type Service = service.Service

type Handler = web.Handler

const (
	EndProcess = domain.END
)

type Order = domain.Order
