package ioc

import (
	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/event"
	attrSvc "github.com/Duke1616/ecmdb/internal/service/attribute"
	"github.com/Duke1616/ecmdb/pkg/mqx"
	"github.com/ecodeclub/mq-api"
)

func InitFieldSecureAttrChangeEventProducer(q mq.MQ) (attrSvc.FieldSecureAttrChangeEventProducer, error) {
	return mqx.NewGeneralProducer[domain.FieldSecureAttrChange](q, event.FieldSecureAttrChangeName)
}

func InitFieldDeleteEventProducer(q mq.MQ) (attrSvc.IFieldDeleteEventProducer, error) {
	return mqx.NewGeneralProducer[domain.FieldDelete](q, event.FIELD_DELETE_EVENT_NAME)
}
