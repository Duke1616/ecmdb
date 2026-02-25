package channel

import (
	"github.com/Duke1616/ecmdb/internal/pkg/notification/provider"
	"github.com/gotomicro/ego/core/elog"
)

type larkCardChannel struct {
	baseChannel
}

func NewLarkCardChannel(builder provider.SelectorBuilder) Channel {
	return &larkCardChannel{
		baseChannel{
			builder: builder,
			logger:  elog.DefaultLogger,
		},
	}
}
