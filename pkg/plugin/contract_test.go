package plugin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalServiceRuntime(t *testing.T) {
	testCases := []struct {
		name       string
		upstream   string
		opts       []RuntimeOption
		wantMode   string
		wantURL    string
		wantHealth string
	}{
		{
			name:       "标准上游地址与健康检查路径",
			upstream:   "http://ssh-plugin:8080/",
			opts:       []RuntimeOption{RuntimeHealthPath("/healthz")},
			wantMode:   types.RuntimeModeExternalService,
			wantURL:    "http://ssh-plugin:8080",
			wantHealth: "/healthz",
		},
		{
			name:       "缺省健康检查路径",
			upstream:   "https://gateway.example.com",
			opts:       nil,
			wantMode:   types.RuntimeModeExternalService,
			wantURL:    "https://gateway.example.com",
			wantHealth: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			def := NewRegistry("ssh", "SSH", ExternalServiceRuntime(tc.upstream, tc.opts...)).
				Action("terminal", "SSH 终端").
				MustDefinition()

			runtime, ok := def.Plugin.Runtime()
			require.True(t, ok)
			assert.Equal(t, tc.wantMode, runtime.Mode)
			assert.Equal(t, tc.wantURL, runtime.Upstream)
			assert.Equal(t, tc.wantHealth, runtime.HealthPath)
		})
	}
}

func TestDefinitionHandler(t *testing.T) {
	testCases := []struct {
		name       string
		method     string
		provider   Provider
		wantCode   int
		assertBody func(t *testing.T, body []byte)
	}{
		{
			name:   "GET 请求正常响应自描述 JSON",
			method: http.MethodGet,
			provider: ProviderFunc(func() (Definition, error) {
				return NewRegistry("ssh", "SSH").
					Action("terminal", "SSH 终端").
					Definition()
			}),
			wantCode: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var def Definition
				err := json.Unmarshal(body, &def)
				require.NoError(t, err)
				assert.Equal(t, "ssh", def.Plugin.UID)
				require.Len(t, def.Plugin.Actions, 1)
				assert.Equal(t, "terminal", def.Plugin.Actions[0].Action)
			},
		},
		{
			name:   "非 GET 请求返回 405 Method Not Allowed",
			method: http.MethodPost,
			provider: ProviderFunc(func() (Definition, error) {
				return Definition{}, nil
			}),
			wantCode: http.StatusMethodNotAllowed,
			assertBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "method not allowed")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, types.WellKnownPath, nil)
			rec := httptest.NewRecorder()
			DefinitionHandler(tc.provider).ServeHTTP(rec, req)

			assert.Equal(t, tc.wantCode, rec.Code)
			tc.assertBody(t, rec.Body.Bytes())
		})
	}
}
