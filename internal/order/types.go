package order

import (
	"github.com/Duke1616/ecmdb/internal/order/internal/domain"
	"github.com/Duke1616/ecmdb/internal/order/internal/grpc"
	"github.com/Duke1616/ecmdb/internal/order/internal/service"
	"github.com/Duke1616/ecmdb/internal/order/internal/web"
)

type Service = service.Service

type RpcServer = grpc.WorkOrderServer

type Handler = web.Handler

const (
	EndProcess    = domain.END
	SystemProvide = domain.SYSTEM
	WechatProvide = domain.WECHAT
)

type Order = domain.Order
