package policy

import (
	"github.com/Duke1616/ecmdb/internal/policy/internal/domain"
	"github.com/Duke1616/ecmdb/internal/policy/internal/grpc"
	"github.com/Duke1616/ecmdb/internal/policy/internal/service"
	"github.com/Duke1616/ecmdb/internal/policy/internal/web"
)

type Handler = web.Handler

type RpcServer = grpc.PolicyServer

type Service = service.Service

type Policy = domain.Policy

type Policies = domain.Policies
