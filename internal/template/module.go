package template

import (
	"github.com/Duke1616/ecmdb/internal/template/internal/event"
)

type Module struct {
	Svc      Service
	c        *event.WechatApprovalCallbackConsumer
	Hdl      *Handler
	GroupHdl *GroupHdl
}
