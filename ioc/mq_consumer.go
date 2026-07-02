package ioc

import (
	"github.com/Duke1616/ecmdb/internal/event"
	resourceEvent "github.com/Duke1616/ecmdb/internal/event/resource"
	resourceSvc "github.com/Duke1616/ecmdb/internal/service/resource"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/ecodeclub/mq-api"
)

func InitFieldSecureAttrChangeConsumer(q mq.MQ, svc resourceSvc.EncryptedSvc, crypto cryptox.Crypto) (*resourceEvent.FieldSecureAttrChangeConsumer, error) {
	consumer, err := q.Consumer(event.FieldSecureAttrChangeName, "field_secure_change")
	if err != nil {
		return nil, err
	}
	return resourceEvent.NewFieldSecureAttrChangeConsumer(consumer, svc, 100, crypto), nil
}

func InitFieldDeleteConsumer(q mq.MQ, svc resourceSvc.Service) (*resourceEvent.FieldDeleteConsumer, error) {
	consumer, err := q.Consumer(event.FIELD_DELETE_EVENT_NAME, "field_delete")
	if err != nil {
		return nil, err
	}
	return resourceEvent.NewFieldDeleteConsumer(consumer, svc), nil
}
