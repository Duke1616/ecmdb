package workflow

import (
	"github.com/Duke1616/ecmdb/internal/workflow/internal/domain"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/service"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/web"
)

type Handler = web.Handler

type Service = service.Service

type Workflow = domain.Workflow

type NotifyMethod = domain.NotifyMethod

// NotifyBinding 暴露 domain.NotifyBinding
type NotifyBinding = domain.NotifyBinding

// NotifyType 暴露 domain.NotifyType
type NotifyType = domain.NotifyType

const (
	NotifyTypeApproval            NotifyType = domain.NotifyTypeApproval
	NotifyTypeCC                  NotifyType = domain.NotifyTypeCC
	NotifyTypeProgress            NotifyType = domain.NotifyTypeProgress
	NotifyTypeProgressImageResult NotifyType = domain.NotifyTypeProgressImageResult
	NotifyTypeRevoke              NotifyType = domain.NotifyTypeRevoke
)
