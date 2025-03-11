package worker

import (
	"github.com/Duke1616/ecmdb/internal/worker/internal/domain"
	"github.com/Duke1616/ecmdb/internal/worker/internal/service"
	"github.com/Duke1616/ecmdb/internal/worker/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Execute = domain.Execute

type Worker = domain.Worker

const (
	RUNNING  = domain.RUNNING
	STOPPING = domain.STOPPING
	OFFLINE  = domain.OFFLINE
)
