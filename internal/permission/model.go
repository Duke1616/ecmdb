package permission

import "github.com/Duke1616/ecmdb/internal/permission/internal/event"

type Module struct {
	Hdl *Handler
	Svc Service
	c   *event.MenuChangeEventConsumer
}
