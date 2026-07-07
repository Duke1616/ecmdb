package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	RuntimeModeBuiltin         = "builtin"
	RuntimeModeExternalService = "external-service"

	MetaDescription = "description"
	MetaRuntime     = "runtime"

	HeaderPluginID = "X-ECMDB-Plugin-ID"

	WellKnownPath = "/.well-known/ecmdb-plugin"

	DefinitionFetchTimeout = 3 * time.Second
)

// Provider 是插件实现方需要提供的最小契约。
type Provider interface {
	Definition() (Definition, error)
}

type ProviderFunc func() (Definition, error)

func (fn ProviderFunc) Definition() (Definition, error) {
	return fn()
}

// ContextResolver 描述 ECMDB Core 暴露给插件后端的资源上下文解析能力。
//
// 插件后端只传 plugin_id、action、resource_id 和必要参数，敏感字段由 ECMDB 后端解析，
// 避免 password/private_key 等内容经过浏览器。
type ContextResolver interface {
	ResolveActionContext(ctx context.Context, req ResolveRequest) (ActionContext, error)
}

type RuntimeSpec struct {
	Mode       string `json:"mode"`
	Upstream   string `json:"upstream,omitempty"`
	HealthPath string `json:"health_path,omitempty"`
}

type RuntimeOption func(*RuntimeSpec)

func RuntimeHealthPath(path string) RuntimeOption {
	return func(spec *RuntimeSpec) {
		spec.HealthPath = strings.TrimSpace(path)
	}
}

func Description(value string) Option {
	return func(p *Plugin) {
		if p.Meta == nil {
			p.Meta = make(map[string]any)
		}
		p.Meta[MetaDescription] = strings.TrimSpace(value)
	}
}

func BuiltinRuntime() Option {
	return Runtime(RuntimeSpec{Mode: RuntimeModeBuiltin})
}

func ExternalServiceRuntime(upstream string, opts ...RuntimeOption) Option {
	spec := RuntimeSpec{
		Mode:     RuntimeModeExternalService,
		Upstream: strings.TrimRight(strings.TrimSpace(upstream), "/"),
	}
	for _, opt := range opts {
		opt(&spec)
	}
	return Runtime(spec)
}

func Runtime(spec RuntimeSpec) Option {
	return func(p *Plugin) {
		p.SetRuntime(spec)
	}
}

func (p *Plugin) SetRuntime(spec RuntimeSpec) {
	if p.Meta == nil {
		p.Meta = make(map[string]any)
	}
	p.Meta[MetaRuntime] = spec
}

func (p Plugin) Runtime() (RuntimeSpec, bool) {
	value, ok := p.Meta[MetaRuntime]
	if !ok || value == nil {
		return RuntimeSpec{}, false
	}

	switch spec := value.(type) {
	case RuntimeSpec:
		return spec, true
	case *RuntimeSpec:
		if spec == nil {
			return RuntimeSpec{}, false
		}
		return *spec, true
	case map[string]any:
		return runtimeFromMap(spec)
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return RuntimeSpec{}, false
		}
		var decoded RuntimeSpec
		if err = json.Unmarshal(data, &decoded); err != nil {
			return RuntimeSpec{}, false
		}
		return decoded, true
	}
}

func runtimeFromMap(src map[string]any) (RuntimeSpec, bool) {
	spec := RuntimeSpec{
		Mode:       stringFromMap(src, "mode"),
		Upstream:   stringFromMap(src, "upstream"),
		HealthPath: stringFromMap(src, "health_path"),
	}
	return spec, spec.Mode != "" || spec.Upstream != "" || spec.HealthPath != ""
}

func DefinitionURL(upstream string) (string, error) {
	upstream = strings.TrimRight(strings.TrimSpace(upstream), "/")
	if upstream == "" {
		return "", fmt.Errorf("upstream 地址不能为空")
	}
	return upstream + WellKnownPath, nil
}

func FetchDefinition(ctx context.Context, upstream string) (Definition, error) {
	endpoint, err := DefinitionURL(upstream)
	if err != nil {
		return Definition{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, DefinitionFetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Definition{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Definition{}, fmt.Errorf("读取插件自描述失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Definition{}, fmt.Errorf("读取插件自描述返回非 200 状态码: %d", resp.StatusCode)
	}

	var def Definition
	if err = json.NewDecoder(resp.Body).Decode(&def); err != nil {
		return Definition{}, fmt.Errorf("解析插件自描述失败: %w", err)
	}
	return def, nil
}

func stringFromMap(src map[string]any, key string) string {
	val, ok := src[key]
	if !ok || val == nil {
		return ""
	}
	return fmt.Sprint(val)
}

func DefinitionHandler(provider Provider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		def, err := provider.Definition()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(def)
	})
}

func MountWellKnown(mux interface {
	Handle(pattern string, handler http.Handler)
}, provider Provider) {
	mux.Handle(WellKnownPath, DefinitionHandler(provider))
}
