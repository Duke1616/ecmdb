package channel

import (
	"github.com/Duke1616/ecmdb/internal/pkg/notification/provider"
	"github.com/gotomicro/ego/core/elog"
)

type larkTextChannel struct {
	baseChannel
}

func NewLarkTextChannel(builder provider.SelectorBuilder) Channel {
	return &larkTextChannel{
		baseChannel{
			builder: builder,
			logger:  elog.DefaultLogger,
		},
	}
}
