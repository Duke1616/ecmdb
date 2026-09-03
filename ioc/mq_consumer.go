package ioc

import (
	"github.com/Duke1616/ecmdb/internal/event"
	resourceEvent "github.com/Duke1616/ecmdb/internal/event/resource"
	resourceSvc "github.com/Duke1616/ecmdb/internal/service/resource"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/Duke1616/ecmdb/pkg/mqx"
	"github.com/ecodeclub/mq-api"
)

func InitFieldSecureAttrChangeConsumer(q mq.MQ, svc resourceSvc.EncryptedSvc, crypto cryptox.Crypto) (*resourceEvent.FieldSecureAttrChangeConsumer, error) {
	consumer := mqx.NewResilientConsumer(q, event.FieldSecureAttrChangeEventName, "field_secure_change")
	return resourceEvent.NewFieldSecureAttrChangeConsumer(consumer, svc, 100, crypto), nil
}

func InitFieldDeleteConsumer(q mq.MQ, svc resourceSvc.Service) (*resourceEvent.FieldDeleteConsumer, error) {
	consumer := mqx.NewResilientConsumer(q, event.FieldDeleteEventName, "field_delete")
	return resourceEvent.NewFieldDeleteConsumer(consumer, svc), nil
}
