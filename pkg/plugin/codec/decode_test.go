package codec

import (
	"testing"
	"time"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEndpoint struct {
	Host     string `plugin:"host,required"`
	Port     int    `plugin:"port,default=22"`
	Username string `plugin:"username,required"`
}

type testTarget struct {
	testEndpoint
	Gateways []testEndpoint `plugin:"gateways"`
}

func makeActionCtx(inputs map[string]types.ResolvedInput) types.ActionContext {
	return types.ActionContext{Inputs: inputs}
}

func singleInput(name, cardinality string, resources ...types.ResolvedResource) map[string]types.ResolvedInput {
	return map[string]types.ResolvedInput{
		name: {Name: name, Cardinality: cardinality, Resources: resources},
	}
}

func TestInputOne(t *testing.T) {
	testCases := []struct {
		name       string
		ctx        types.ActionContext
		inputName  string
		assertVal  func(t *testing.T, val testTarget)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "解码成功_嵌套结构与默认值",
			ctx: makeActionCtx(singleInput("target", types.CardinalityOne, types.ResolvedResource{
				Fields: map[string]any{"host": "10.0.0.1", "username": "root"},
				Children: map[string]types.ResolvedInput{
					"gateways": {
						Name:        "gateways",
						Cardinality: types.CardinalityMany,
						Resources: []types.ResolvedResource{
							{Fields: map[string]any{"host": "10.0.0.10", "port": 2222, "username": "jump"}},
						},
					},
				},
			})),
			inputName: "target",
			assertVal: func(t *testing.T, val testTarget) {
				assert.Equal(t, "10.0.0.1", val.Host)
				assert.Equal(t, 22, val.Port)
				assert.Equal(t, "root", val.Username)
				require.Len(t, val.Gateways, 1)
				assert.Equal(t, "10.0.0.10", val.Gateways[0].Host)
				assert.Equal(t, 2222, val.Gateways[0].Port)
				assert.Equal(t, "jump", val.Gateways[0].Username)
			},
		},
		{
			name:       "input 名称不存在返回错误",
			ctx:        makeActionCtx(map[string]types.ResolvedInput{}),
			inputName:  "target",
			wantErr:    true,
			wantErrMsg: "not found",
		},
		{
			name:       "input 资源列表为空返回错误",
			ctx:        makeActionCtx(singleInput("target", types.CardinalityOne)),
			inputName:  "target",
			wantErr:    true,
			wantErrMsg: "not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := InputOne[testTarget](tc.ctx, tc.inputName)
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tc.wantErrMsg)
				}
				return
			}
			require.NoError(t, err)
			tc.assertVal(t, got)
		})
	}
}

func TestDecodeResource(t *testing.T) {
	type wantResult struct {
		host     string
		port     int
		username string
	}

	testCases := []struct {
		name      string
		resource  types.ResolvedResource
		wantErr   bool
		wantValue wantResult
	}{
		{
			name: "全字段正常解码",
			resource: types.ResolvedResource{
				Fields: map[string]any{"host": "192.168.1.1", "port": 8022, "username": "admin"},
			},
			wantValue: wantResult{host: "192.168.1.1", port: 8022, username: "admin"},
		},
		{
			name: "port 缺失时使用默认值 22",
			resource: types.ResolvedResource{
				Fields: map[string]any{"host": "192.168.1.2", "username": "root"},
			},
			wantValue: wantResult{host: "192.168.1.2", port: 22, username: "root"},
		},
		{
			name: "port 为字符串类型自动类型转换",
			resource: types.ResolvedResource{
				Fields: map[string]any{"host": "192.168.1.3", "port": "2222", "username": "ops"},
			},
			wantValue: wantResult{host: "192.168.1.3", port: 2222, username: "ops"},
		},
		{
			name:     "required 字段 host 缺失返回错误",
			resource: types.ResolvedResource{Fields: map[string]any{"username": "root"}},
			wantErr:  true,
		},
		{
			name:     "required 字段 username 缺失返回错误",
			resource: types.ResolvedResource{Fields: map[string]any{"host": "10.0.0.1"}},
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DecodeResource[testEndpoint](tc.resource)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantValue.host, got.Host)
			assert.Equal(t, tc.wantValue.port, got.Port)
			assert.Equal(t, tc.wantValue.username, got.Username)
		})
	}
}

type testRichResource struct {
	PrivateKey []byte    `plugin:"private_key"`
	CreatedAt  time.Time `plugin:"created_at"`
	Enabled    bool      `plugin:"enabled"`
	Score      float64   `plugin:"score"`
}

func TestDecodeResource_RichTypes(t *testing.T) {
	res := types.ResolvedResource{
		Fields: map[string]any{
			"private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA0...",
			"created_at":  "2026-09-17 12:30:00",
			"enabled":     "true",
			"score":       "98.5",
		},
	}

	got, err := DecodeResource[testRichResource](res)
	require.NoError(t, err)
	assert.Equal(t, []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA0..."), got.PrivateKey)
	assert.Equal(t, 2026, got.CreatedAt.Year())
	assert.Equal(t, time.September, got.CreatedAt.Month())
	assert.Equal(t, 17, got.CreatedAt.Day())
	assert.True(t, got.Enabled)
	assert.Equal(t, 98.5, got.Score)
}
