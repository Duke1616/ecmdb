package rota

import (
	"github.com/Duke1616/ecmdb/internal/rota/internal/grpc"
	"github.com/Duke1616/ecmdb/internal/rota/internal/service"
	"github.com/Duke1616/ecmdb/internal/rota/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type RpcServer = grpc.RotaServer
