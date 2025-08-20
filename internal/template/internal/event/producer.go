package event

import (
	"context"

	"github.com/Duke1616/ecmdb/pkg/mqx"
	"github.com/ecodeclub/mq-api"
	"github.com/xen0n/go-workwx"
)

type WechatOrderEventProducer interface {
	Produce(ctx context.Context, evt *workwx.OAApprovalDetail) error
}

func NewWechatOrderEventProducer(q mq.MQ) (WechatOrderEventProducer, error) {
	return mqx.NewGeneralProducer[*workwx.OAApprovalDetail](q, WechatOrderEventName)
}
