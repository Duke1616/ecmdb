package engine

import (
	"github.com/Duke1616/ecmdb/internal/engine/internal/domain"
	"github.com/Duke1616/ecmdb/internal/engine/internal/service"
	"github.com/Duke1616/ecmdb/internal/engine/internal/web"
)

type Service = service.Service

type Handler = web.Handler

type Instance = domain.Instance
