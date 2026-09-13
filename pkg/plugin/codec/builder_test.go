package codec

import (
	"testing"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEndpointSubGateway struct {
	IP    string `plugin:"ip,required"`
	Level int    `plugin:"level"`
}

type testEndpointGateway struct {
	IP          string                    `plugin:"ip,required"`
	Port        int                       `plugin:"port,field=ssh_port"`
	SubGateways []testEndpointSubGateway `plugin:"sub_gateways,model=sub_gateway,in=run"`
}

type testEndpointHost struct {
	IP       string                `plugin:"ip,required"`
	Username string                `plugin:"username"`
	Gateways []testEndpointGateway `plugin:"gateways,model=gateway,out=default"`
}

func TestBuildCenterSpec(t *testing.T) {
	testCases := []struct {
		name       string
		build      func() (types.ResourceSpec, error)
		assertSpec func(t *testing.T, spec types.ResourceSpec)
		wantErr    bool
	}{
		{
			name: "多级关联中心模型构建",
			build: func() (types.ResourceSpec, error) {
				return BuildCenterSpec[testEndpointHost]("target", "host")
			},
			assertSpec: func(t *testing.T, spec types.ResourceSpec) {
				assert.Equal(t, "target", spec.Name)
				assert.Equal(t, "host", spec.ModelUID)
				assert.Equal(t, "ip", spec.Fields["ip"])
				assert.Equal(t, "username", spec.Fields["username"])

				require.Len(t, spec.Children, 1)
				gateways := spec.Children[0]
				assert.Equal(t, "gateways", gateways.Name)
				assert.Equal(t, types.CardinalityMany, gateways.Cardinality)
				assert.Equal(t, "gateway", gateways.ModelUID)
				assert.Equal(t, types.RelationTypeDefault, gateways.RelationType)
				assert.Equal(t, types.DirectionToTarget, gateways.Direction)

				require.Len(t, gateways.Children, 1)
				sub := gateways.Children[0]
				assert.Equal(t, "sub_gateways", sub.Name)
				assert.Equal(t, "sub_gateway", sub.ModelUID)
				assert.Equal(t, types.RelationTypeRun, sub.RelationType)
				assert.Equal(t, types.DirectionToSource, sub.Direction)
			},
		},
		{
			name: "非结构体类型报错",
			build: func() (types.ResourceSpec, error) {
				return BuildCenterSpec[string]("target", "host")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			spec, err := tc.build()
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			tc.assertSpec(t, spec)
		})
	}
}

func (testEndpointHost) ModelName() string  { return "主机资产" }
func (testEndpointHost) ModelGroup() string { return "计算资源" }

func TestInspectTarget(t *testing.T) {
	testCases := []struct {
		name       string
		inspect    func() (TargetMeta, error)
		assertMeta func(t *testing.T, meta TargetMeta)
	}{
		{
			name: "单次内省同时输出 Schema 与 Spec",
			inspect: func() (TargetMeta, error) {
				return InspectTarget[testEndpointHost]("target", "host")
			},
			assertMeta: func(t *testing.T, meta TargetMeta) {
				assert.Equal(t, "主机资产", meta.Spec.ModelName)
				assert.Equal(t, "计算资源", meta.Spec.GroupName)

				require.Len(t, meta.Schema.Models, 3)
				require.Len(t, meta.Schema.ModelRelations, 2)
				require.Len(t, meta.Schema.ModelGroups, 1)
				assert.Equal(t, "计算资源", meta.Schema.ModelGroups[0].Name)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			meta, err := tc.inspect()
			require.NoError(t, err)
			tc.assertMeta(t, meta)
		})
	}
}
