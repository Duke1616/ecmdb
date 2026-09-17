package codec

import (
	"testing"
	"time"

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

type testAuthType string

func (testAuthType) EnumOptions() []string {
	return []string{"passwd", "publickey", "passphrase"}
}

type testEntityWithEnum struct {
	AuthType     testAuthType `plugin:"auth_type,label=认证方式"`
	ExplicitList string       `plugin:"explicit_list,type=list,options=opt1|opt2"`
	PrivateKey   []byte       `plugin:"private_key,label=私钥凭证"`
	CreatedAt    time.Time    `plugin:"created_at,label=创建时间"`
	Enabled      bool         `plugin:"enabled,label=是否启用"`
	Port         int          `plugin:"port,label=端口号"`
}

type testSecureExplicitEntity struct {
	NormalField     string `plugin:"normal,label=普通字段"`
	ExplicitSecure  string `plugin:"custom_key,secure,label=自定义敏感字段"`
	ExplicitEncrypt string `plugin:"cipher_text,encrypt=true,label=加密字段"`
	KeywordOverride string `plugin:"token_status,secure=false,label=令牌状态"`
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
			name: "枚举自描述接口推导与 Tag options 支持",
			derive: func() (types.Schema, error) {
				return DeriveSchema[testEntityWithEnum]("enum_entity")
			},
			assertSchema: func(t *testing.T, s types.Schema) {
				require.Len(t, s.Models, 1)
				fields := s.Models[0].AttributeGroups[0].Fields
				require.Len(t, fields, 6)

				// 1. 验证 IEnum 接口自动推导出 list 与 options
				authField, ok := lo.Find(fields, func(a types.Attribute) bool { return a.UID == "auth_type" })
				require.True(t, ok)
				assert.Equal(t, "list", authField.Type)
				assert.Equal(t, []string{"passwd", "publickey", "passphrase"}, authField.Option)

				// 2. 验证通过 Tag 的 type=list,options= 声明
				listField, ok := lo.Find(fields, func(a types.Attribute) bool { return a.UID == "explicit_list" })
				require.True(t, ok)
				assert.Equal(t, "list", listField.Type)
				assert.Equal(t, []string{"opt1", "opt2"}, listField.Option)

				// 3. 验证 []byte 自动推导为 multiline，且敏感字段 secure=true
				pkField, ok := lo.Find(fields, func(a types.Attribute) bool { return a.UID == "private_key" })
				require.True(t, ok)
				assert.Equal(t, "multiline", pkField.Type)
				assert.True(t, pkField.Secure)

				// 4. 验证 time.Time 自动推导为 datetime
				timeField, ok := lo.Find(fields, func(a types.Attribute) bool { return a.UID == "created_at" })
				require.True(t, ok)
				assert.Equal(t, "datetime", timeField.Type)

				// 5. 验证 bool 自动推导为 boolean
				boolField, ok := lo.Find(fields, func(a types.Attribute) bool { return a.UID == "enabled" })
				require.True(t, ok)
				assert.Equal(t, "boolean", boolField.Type)

				// 6. 验证 int 自动推导为 number
				numField, ok := lo.Find(fields, func(a types.Attribute) bool { return a.UID == "port" })
				require.True(t, ok)
				assert.Equal(t, "number", numField.Type)
			},
		},
		{
			name: "显式 secure 与 encrypt 标记及关键字覆盖",
			derive: func() (types.Schema, error) {
				return DeriveSchema[testSecureExplicitEntity]("secure_entity")
			},
			assertSchema: func(t *testing.T, s types.Schema) {
				require.Len(t, s.Models, 1)
				fields := s.Models[0].AttributeGroups[0].Fields
				require.Len(t, fields, 4)

				normal, ok := lo.Find(fields, func(a types.Attribute) bool { return a.UID == "normal" })
				require.True(t, ok)
				assert.False(t, normal.Secure)
				assert.True(t, normal.Display)

				// 1. 显式 secure 修饰符
				sec, ok := lo.Find(fields, func(a types.Attribute) bool { return a.UID == "custom_key" })
				require.True(t, ok)
				assert.True(t, sec.Secure)
				assert.False(t, sec.Display)

				// 2. 显式 encrypt=true 修饰符
				enc, ok := lo.Find(fields, func(a types.Attribute) bool { return a.UID == "cipher_text" })
				require.True(t, ok)
				assert.True(t, enc.Secure)
				assert.False(t, enc.Display)

				// 3. 显式 secure=false 覆盖关键字推导（token 关键字被显式关闭）
				override, ok := lo.Find(fields, func(a types.Attribute) bool { return a.UID == "token_status" })
				require.True(t, ok)
				assert.False(t, override.Secure)
				assert.True(t, override.Display)
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
