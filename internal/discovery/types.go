package discovery

import (
	"github.com/Duke1616/ecmdb/internal/discovery/internal/domain"
	"github.com/Duke1616/ecmdb/internal/discovery/internal/service"
	"github.com/Duke1616/ecmdb/internal/discovery/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Discovery = domain.Discovery
