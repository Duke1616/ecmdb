package plugin

import (
	"context"
	"fmt"
	"strings"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
)

// Provider 是插件实现方需要提供的最小契约
type Provider interface {
	Definition() (Definition, error)
}

type ProviderFunc func() (Definition, error)

func (fn ProviderFunc) Definition() (Definition, error) {
	return fn()
}

// ContextResolver 描述 ECMDB Core 暴露给插件后端的资源上下文解析能力
type ContextResolver interface {
	ResolveActionContext(ctx context.Context, req types.ResolveRequest) (types.ActionContext, error)
}

type RuntimeOption func(*types.RuntimeSpec)

func RuntimeHealthPath(path string) RuntimeOption {
	return func(spec *types.RuntimeSpec) {
		spec.HealthPath = strings.TrimSpace(path)
	}
}

func Description(value string) Option {
	return func(p *types.Plugin) {
		if p.Meta == nil {
			p.Meta = make(map[string]any)
		}
		p.Meta["description"] = strings.TrimSpace(value)
	}
}

func BuiltinRuntime() Option {
	return Runtime(types.RuntimeSpec{Mode: types.RuntimeModeBuiltin})
}

func ExternalServiceRuntime(upstream string, opts ...RuntimeOption) Option {
	spec := types.RuntimeSpec{
		Mode:     types.RuntimeModeExternalService,
		Upstream: strings.TrimRight(strings.TrimSpace(upstream), "/"),
	}
	for _, opt := range opts {
		opt(&spec)
	}
	return Runtime(spec)
}

func Runtime(spec types.RuntimeSpec) Option {
	return func(p *types.Plugin) {
		p.SetRuntime(spec)
	}
}

func DefinitionURL(upstream string) (string, error) {
	upstream = strings.TrimRight(strings.TrimSpace(upstream), "/")
	if upstream == "" {
		return "", fmt.Errorf("upstream 地址不能为空")
	}
	return upstream + types.WellKnownPath, nil
}
