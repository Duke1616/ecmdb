package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
)

// FetchDefinition 从外部插件服务的 Well-Known 端点拉取 Definition 自描述
func FetchDefinition(ctx context.Context, upstream string) (Definition, error) {
	endpoint, err := DefinitionURL(upstream)
	if err != nil {
		return Definition{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
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

// DefinitionHandler 将 Provider 暴露为标准 HTTP Handler，响应 Well-Known 自描述请求
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

// MountWellKnown 将 DefinitionHandler 挂载到标准 Well-Known 路径
func MountWellKnown(mux interface {
	Handle(pattern string, handler http.Handler)
}, provider Provider) {
	mux.Handle(types.WellKnownPath, DefinitionHandler(provider))
}
