package codec

import (
	"testing"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testDeriveSubGateway struct {
	IP    string `plugin:"ip,required"`
	Level int    `plugin:"level"`
}

type testDeriveGateway struct {
	IP          string                 `plugin:"ip,required"`
	Port        int                    `plugin:"port,field=ssh_port"`
	SubGateways []testDeriveSubGateway `plugin:"sub_gateways,model=sub_gateway,in=run"`
}

type testDeriveHost struct {
	IP       string              `plugin:"ip,required"`
	Username string              `plugin:"username"`
	Password string              `plugin:"password"` // 敏感字段，预期 secure=true, display=false
	Gateways []testDeriveGateway `plugin:"gateways,model=gateway,out=default"`
}

type testSelfDescribedEntity struct {
	Name string `plugin:"name"`
}

func (testSelfDescribedEntity) ModelName() string {
	return "自描述实体"
}

func (testSelfDescribedEntity) ModelGroup() string {
	return "自描述组"
}

type testCompactEntity struct {
	Name string `plugin:"name"`
}

func (testCompactEntity) DescribeModel() (string, string) {
	return "紧凑自描述模型", "核心凭据组"
}

func TestDeriveSchema(t *testing.T) {
	testCases := []struct {
		name         string
		derive       func() (types.Schema, error)
		assertSchema func(t *testing.T, s types.Schema)
	}{
		{
			name: "多模型递归推导与安全字段识别",
			derive: func() (types.Schema, error) {
				return DeriveSchema[testDeriveHost]("host")
			},
			assertSchema: func(t *testing.T, s types.Schema) {
				require.Len(t, s.Models, 3)

				hostModel, ok := lo.Find(s.Models, func(m types.ModelSpec) bool { return m.UID == "host" })
				require.True(t, ok)
				fields := hostModel.AttributeGroups[0].Fields
				require.Len(t, fields, 3)

				// 验证 password 为加密敏感字段
				pwdField, ok := lo.Find(fields, func(a types.Attribute) bool { return a.UID == "password" })
				require.True(t, ok)
				assert.True(t, pwdField.Secure)
				assert.False(t, pwdField.Display)

				// 验证关联关系
				require.Len(t, s.ModelRelations, 2)
				assert.True(t, lo.SomeBy(s.ModelRelations, func(r types.ModelRelation) bool {
					return r.SourceModelUID == "host" && r.TargetModelUID == "gateway" && r.RelationTypeUID == types.RelationTypeDefault
				}))
				assert.True(t, lo.SomeBy(s.ModelRelations, func(r types.ModelRelation) bool {
					return r.SourceModelUID == "sub_gateway" && r.TargetModelUID == "gateway" && r.RelationTypeUID == types.RelationTypeRun
				}))
			},
		},
		{
			name: "结构体接口自描述推导",
			derive: func() (types.Schema, error) {
				return DeriveSchema[testSelfDescribedEntity]("custom_entity")
			},
			assertSchema: func(t *testing.T, s types.Schema) {
				require.Len(t, s.Models, 1)
				assert.Equal(t, "custom_entity", s.Models[0].UID)
				assert.Equal(t, "自描述实体", s.Models[0].Name)
				assert.Equal(t, "自描述组", s.Models[0].GroupName)
				require.Len(t, s.ModelGroups, 1)
				assert.Equal(t, "自描述组", s.ModelGroups[0].Name)
			},
		},
		{
			name: "紧凑结构体接口自描述推导 (DescribeModel)",
			derive: func() (types.Schema, error) {
				return DeriveSchema[testCompactEntity]("compact_entity")
			},
			assertSchema: func(t *testing.T, s types.Schema) {
				require.Len(t, s.Models, 1)
				assert.Equal(t, "compact_entity", s.Models[0].UID)
				assert.Equal(t, "紧凑自描述模型", s.Models[0].Name)
				assert.Equal(t, "核心凭据组", s.Models[0].GroupName)
				require.Len(t, s.ModelGroups, 1)
				assert.Equal(t, "核心凭据组", s.ModelGroups[0].Name)
			},
		},
		{
			name: "通过结构体内嵌 Tag 自描述推导",
			derive: func() (types.Schema, error) {
				type embeddedTagEntity struct {
					testSelfDescribedEntity `plugin:",label=内嵌模型,group=内嵌组"`
				}
				return DeriveSchema[embeddedTagEntity]("embedded_entity")
			},
			assertSchema: func(t *testing.T, s types.Schema) {
				require.Len(t, s.Models, 1)
				assert.Equal(t, "embedded_entity", s.Models[0].UID)
				assert.Equal(t, "内嵌模型", s.Models[0].Name)
				assert.Equal(t, "内嵌组", s.Models[0].GroupName)
				require.Len(t, s.ModelGroups, 1)
				assert.Equal(t, "内嵌组", s.ModelGroups[0].Name)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := tc.derive()
			require.NoError(t, err)
			tc.assertSchema(t, s)
		})
	}
}
