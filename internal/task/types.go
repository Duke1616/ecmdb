package task

import (
	"github.com/Duke1616/ecmdb/internal/task/internal/job"
	"github.com/Duke1616/ecmdb/internal/task/internal/service"
	"github.com/Duke1616/ecmdb/internal/task/internal/web"
)

type Service = service.Service

type Handler = web.Handler

type StartTaskJob = job.StartTaskJob

type PassProcessTaskJob = job.PassProcessTaskJob

type RecoveryTaskJob = job.RecoveryTaskJob
