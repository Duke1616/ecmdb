package template

import (
	"github.com/Duke1616/ecmdb/internal/template/internal/domain"
	"github.com/Duke1616/ecmdb/internal/template/internal/service"
	"github.com/Duke1616/ecmdb/internal/template/internal/web"
)

type Handler = web.Handler

type GroupHdl = web.GroupHandler

type Service = service.Service

type Template = domain.Template
