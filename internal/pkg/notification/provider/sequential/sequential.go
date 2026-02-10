package sequential

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/ecmdb/internal/pkg/notification/provider"
)

var (
	_ provider.Selector        = (*selector)(nil)
	_ provider.SelectorBuilder = (*SelectorBuilder)(nil)
)

// selector 供应商顺序选择器
type selector struct {
	idx       int
	providers []provider.Provider
}

func (r *selector) Next(_ context.Context, _ notification.Notification) (provider.Provider, error) {
	if len(r.providers) == r.idx {
		return nil, fmt.Errorf("%s", "无可用渠道")
	}

	p := r.providers[r.idx]
	r.idx++
	return p, nil
}

type SelectorBuilder struct {
	providers []provider.Provider
}

func NewSelectorBuilder(providers []provider.Provider) *SelectorBuilder {
	return &SelectorBuilder{providers: providers}
}

func (s *SelectorBuilder) Build() (provider.Selector, error) {
	return &selector{
		providers: s.providers,
	}, nil
}
