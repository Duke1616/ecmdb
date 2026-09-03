package codec

import (
	"testing"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
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
	t.Run("解码成功_嵌套结构与默认值", func(t *testing.T) {
		ctx := makeActionCtx(singleInput("target", types.CardinalityOne, types.ResolvedResource{
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
		}))

		target, err := InputOne[testTarget](ctx, "target")
		if err != nil {
			t.Fatalf("InputOne() error = %v", err)
		}
		if target.Host != "10.0.0.1" {
			t.Errorf("Host = %s, want 10.0.0.1", target.Host)
		}
		if target.Port != 22 {
			t.Errorf("Port (default) = %d, want 22", target.Port)
		}
		if target.Username != "root" {
			t.Errorf("Username = %s, want root", target.Username)
		}
		if len(target.Gateways) != 1 {
			t.Fatalf("len(Gateways) = %d, want 1", len(target.Gateways))
		}
		if target.Gateways[0].Host != "10.0.0.10" || target.Gateways[0].Port != 2222 {
			t.Errorf("Gateway = %#v", target.Gateways[0])
		}
	})

	t.Run("input 名称不存在_返回错误", func(t *testing.T) {
		ctx := makeActionCtx(map[string]types.ResolvedInput{})
		_, err := InputOne[testTarget](ctx, "target")
		if err == nil {
			t.Fatal("expected error for missing input, got nil")
		}
	})

	t.Run("input 资源列表为空_返回错误", func(t *testing.T) {
		ctx := makeActionCtx(singleInput("target", types.CardinalityOne))
		_, err := InputOne[testTarget](ctx, "target")
		if err == nil {
			t.Fatal("expected error for empty resources, got nil")
		}
	})
}

func TestDecodeResource(t *testing.T) {
	type wantResult struct {
		host     string
		port     int
		username string
	}

	cases := []struct {
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
			name: "port 为字符串类型自动转换",
			resource: types.ResolvedResource{
				Fields: map[string]any{"host": "192.168.1.3", "port": "2222", "username": "ops"},
			},
			wantValue: wantResult{host: "192.168.1.3", port: 2222, username: "ops"},
		},
		{
			name:    "required 字段 host 缺失_返回错误",
			resource: types.ResolvedResource{Fields: map[string]any{"username": "root"}},
			wantErr: true,
		},
		{
			name:    "required 字段 username 缺失_返回错误",
			resource: types.ResolvedResource{Fields: map[string]any{"host": "10.0.0.1"}},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DecodeResource[testEndpoint](tc.resource)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil; result = %#v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("DecodeResource() error = %v", err)
			}
			if got.Host != tc.wantValue.host {
				t.Errorf("Host = %s, want %s", got.Host, tc.wantValue.host)
			}
			if got.Port != tc.wantValue.port {
				t.Errorf("Port = %d, want %d", got.Port, tc.wantValue.port)
			}
			if got.Username != tc.wantValue.username {
				t.Errorf("Username = %s, want %s", got.Username, tc.wantValue.username)
			}
		})
	}
}
