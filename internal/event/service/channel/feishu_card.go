package channel

import (
	"github.com/Duke1616/ecmdb/internal/event/service/provider"
)

type feishuCardChannel struct {
	baseChannel
}

func NewFeishuCardChannel(builder provider.SelectorBuilder) Channel {
	return &feishuCardChannel{
		baseChannel{
			builder: builder,
		},
	}
}
