package runner

import (
	"github.com/Duke1616/ecmdb/internal/runner/internal/domain"
	"github.com/Duke1616/ecmdb/internal/runner/internal/service"
	"github.com/Duke1616/ecmdb/internal/runner/internal/web"
)

type Service = service.Service

type Handler = web.Handler

type Runner = domain.Runner

type Variables = domain.Variables
