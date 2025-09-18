package resource

import "github.com/Duke1616/ecmdb/internal/resource/internal/event"

type Module struct {
	Svc          Service
	EncryptedSvc EncryptedSvc
	Hdl          *Handler
	c            *event.FieldSecureAttrChangeConsumer
}
