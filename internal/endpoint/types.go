package endpoint

import (
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/domain"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/grpc"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/service"
	"github.com/Duke1616/ecmdb/internal/endpoint/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Endpoint = domain.Endpoint

type RpcServer = grpc.EndpointServer
