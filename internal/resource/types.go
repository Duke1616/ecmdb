package resource

import (
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/internal/resource/internal/service"
	"github.com/Duke1616/ecmdb/internal/resource/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Resource = domain.Resource

type EncryptedSvc = service.EncryptedSvc
