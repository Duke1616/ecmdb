package plugin

import (
	"testing"

	"github.com/Duke1616/ecmdb/pkg/plugin/graph"
	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testGateway struct {
	IP   string `plugin:"ip,required"`
	Port int    `plugin:"port"`
}

type testHost struct {
	IP       string        `plugin:"ip,required"`
	Username string        `plugin:"username"`
	Gateways []testGateway `plugin:"gateways,model=gateway,out=default"`
}

type testSelfDescribedTarget struct {
	testHost `plugin:",label=自描述主机,group=计算资产"`
}

type testSimpleTarget struct {
	Name string `plugin:"name,required"`
	IP   string `plugin:"ip,field=host_ip"`
}

func (testSimpleTarget) ModelName() string  { return "简单主机" }
func (testSimpleTarget) ModelGroup() string { return "基础资源" }

func (testHost) ModelName() string  { return "主机资产" }
func (testHost) ModelGroup() string { return "计算资源" }

func TestRegistry_PureActions(t *testing.T) {
	testCases := []struct {
		name       string
		pluginUID  string
		pluginName string
		opts       []Option
		actions    []struct {
			action string
			title  string
			opts   []ActionOption
		}
		assertDef func(t *testing.T, def Definition)
	}{
		{
			name:       "纯动作插件无资产绑定",
			pluginUID:  "builtin.tools",
			pluginName: "通用工具箱",
			opts:       []Option{Type("builtin"), Version("2.0.0")},
			actions: []struct {
				action string
				title  string
				opts   []ActionOption
			}{
				{
					action: "ping",
					title:  "连通性检测",
					opts:   []ActionOption{Icon("play"), Permission("plugin:tools:ping")},
				},
				{
					action: "traceroute",
					title:  "路由追踪",
					opts:   []ActionOption{Icon("route")},
				},
			},
			assertDef: func(t *testing.T, def Definition) {
				assert.Equal(t, "builtin.tools", def.Plugin.UID)
				assert.Equal(t, "builtin", def.Plugin.Type)
				assert.Equal(t, "2.0.0", def.Plugin.Version)
				require.Len(t, def.Plugin.Actions, 2)
				assert.Equal(t, "ping", def.Plugin.Actions[0].Action)
				assert.Equal(t, "plugin:tools:ping", def.Plugin.Actions[0].Permission)
				assert.Empty(t, def.Bindings)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reg := NewRegistry(tc.pluginUID, tc.pluginName, tc.opts...)
			for _, act := range tc.actions {
				reg.Action(act.action, act.title, act.opts...)
			}
			def := reg.MustDefinition()
			tc.assertDef(t, def)
		})
	}
}

func TestRegistry_TargetDriven(t *testing.T) {
	testCases := []struct {
		name      string
		setup     func() (Definition, error)
		assertDef func(t *testing.T, def Definition)
		wantErr   bool
	}{
		{
			name: "简单单资产受控绑定",
			setup: func() (Definition, error) {
				reg := NewRegistry("builtin.simple", "单资产插件")
				return Target[testSimpleTarget](reg, "simple_host").
					Action("ping", "Ping").
					Definition()
			},
			assertDef: func(t *testing.T, def Definition) {
				require.Len(t, def.Schema.Models, 1)
				assert.Equal(t, "simple_host", def.Schema.Models[0].UID)
				assert.Equal(t, "简单主机", def.Schema.Models[0].Name)
				assert.Equal(t, "基础资源", def.Schema.Models[0].GroupName)
				require.Len(t, def.Bindings, 1)
				assert.Equal(t, "builtin.simple.simple_host", def.Bindings[0].UID)
			},
		},
		{
			name: "复杂目标关联拓扑推导与多动作挂载",
			setup: func() (Definition, error) {
				reg := NewRegistry("builtin.ssh", "SSH插件")
				return Target[testHost](reg, "host").
					Workspace("terminal", "Web Shell", Icon("terminal")).
					Workspace("sftp", "Web Sftp", Icon("folder")).
					Action("reboot", "重启", Icon("restart")).
					Definition()
			},
			assertDef: func(t *testing.T, def Definition) {
				// 1. 验证动作继承 BindingUID
				require.Len(t, def.Plugin.Actions, 3)
				for _, act := range def.Plugin.Actions {
					assert.Equal(t, "builtin.ssh.host", act.BindingUID)
				}

				// 2. 验证多模型自动派生 (testHost + testGateway)
				require.Len(t, def.Schema.Models, 2)
				var hostModel *types.ModelSpec
				for i := range def.Schema.Models {
					if def.Schema.Models[i].UID == "host" {
						hostModel = &def.Schema.Models[i]
					}
				}
				require.NotNil(t, hostModel)
				assert.Equal(t, "主机资产", hostModel.Name)
				assert.Equal(t, "计算资源", hostModel.GroupName)

				// 3. 验证模型间拓扑关系
				require.Len(t, def.Schema.ModelRelations, 1)
				assert.Equal(t, "host", def.Schema.ModelRelations[0].SourceModelUID)
				assert.Equal(t, "gateway", def.Schema.ModelRelations[0].TargetModelUID)
				assert.Equal(t, types.RelationTypeDefault, def.Schema.ModelRelations[0].RelationTypeUID)

				// 4. 验证 Binding 图谱可正确编译
				require.Len(t, def.Bindings, 1)
				binding := def.Bindings[0]
				specs, err := graph.CompileBindingGraph(binding.Graph)
				require.NoError(t, err)
				assert.Equal(t, "host", specs[0].ModelUID)
				assert.Equal(t, types.RelationTypeDefault, specs[0].Children[0].RelationType)
			},
		},
		{
			name: "通过结构体内嵌 Tag 自描述提取元数据",
			setup: func() (Definition, error) {
				reg := NewRegistry("builtin.auto", "自动化插件")
				return Target[testSelfDescribedTarget](reg, "host").
					Workspace("shell", "终端").
					Definition()
			},
			assertDef: func(t *testing.T, def Definition) {
				require.Len(t, def.Schema.Models, 2)
				var hostModel *types.ModelSpec
				for i := range def.Schema.Models {
					if def.Schema.Models[i].UID == "host" {
						hostModel = &def.Schema.Models[i]
					}
				}
				require.NotNil(t, hostModel)
				assert.Equal(t, "自描述主机", hostModel.Name)
				assert.Equal(t, "计算资产", hostModel.GroupName)
			},
		},
		{
			name: "使用 Builder 链式 Model 方法显式覆盖元数据",
			setup: func() (Definition, error) {
				reg := NewRegistry("builtin.override", "覆盖插件")
				return Target[testSimpleTarget](reg, "simple_host").
					Model("显式覆盖主机", "显式分组").
					Action("ping", "Ping").
					Definition()
			},
			assertDef: func(t *testing.T, def Definition) {
				require.Len(t, def.Schema.Models, 1)
				assert.Equal(t, "simple_host", def.Schema.Models[0].UID)
				assert.Equal(t, "显式覆盖主机", def.Schema.Models[0].Name)
				assert.Equal(t, "显式分组", def.Schema.Models[0].GroupName)
				require.Len(t, def.Schema.ModelGroups, 1)
				assert.Equal(t, "显式分组", def.Schema.ModelGroups[0].Name)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			def, err := tc.setup()
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			tc.assertDef(t, def)
		})
	}
}
