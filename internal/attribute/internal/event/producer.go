package event

import (
	"context"

	"github.com/Duke1616/ecmdb/pkg/mqx"
	"github.com/ecodeclub/mq-api"
)

type FieldSecureAttrChangeEventProducer interface {
	Produce(ctx context.Context, evt FieldSecureAttrChange) error
}

//go:generate mockgen -source=./producer.go -package=evtmocks -destination=./mocks/producer.mock.go -typed NewFieldSecureAttrChangeEventProducer
func NewFieldSecureAttrChangeEventProducer(q mq.MQ) (FieldSecureAttrChangeEventProducer, error) {
	return mqx.NewGeneralProducer[FieldSecureAttrChange](q, FieldSecureAttrChangeName)
}
