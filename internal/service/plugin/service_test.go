package plugin

import (
	"context"
	"testing"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/repository"
	pluginmocks "github.com/Duke1616/ecmdb/internal/service/plugin/mocks"
	relationmocks "github.com/Duke1616/ecmdb/internal/service/relation/mocks"
	coreplugin "github.com/Duke1616/ecmdb/pkg/plugin"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCompleteRelationSpec(t *testing.T) {
	testCases := []struct {
		name     string
		base     string
		spec     pluginx.ResourceSpec
		expected string
	}{
		{
			name: "out to target",
			base: "host",
			spec: pluginx.ResourceSpec{
				ModelUID:     "gateway",
				RelationType: "default",
				Direction:    pluginx.DirectionToTarget,
			},
			expected: "host_default_gateway",
		},
		{
			name: "in from source",
			base: "host",
			spec: pluginx.ResourceSpec{
				ModelUID:     "gateway",
				RelationType: "default",
				Direction:    pluginx.DirectionToSource,
			},
			expected: "gateway_default_host",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildRelationName(tc.base, tc.spec)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestResolveActionContext(t *testing.T) {
	testCases := []struct {
		name        string
		mock        func(ctrl *gomock.Controller) (repository.PluginRepository, IInputResolver)
		req         pluginx.ResolveRequest
		wantErr     bool
		errContains string
		validate    func(t *testing.T, actCtx pluginx.ActionContext)
	}{
		{
			name: "成功解析动作上下文并补全字段",
			mock: func(ctrl *gomock.Controller) (repository.PluginRepository, IInputResolver) {
				repo := pluginmocks.NewMockPluginRepository(ctrl)
				resolver := pluginmocks.NewMockIInputResolver(ctrl)

				repo.EXPECT().GetPlugin(gomock.Any(), "builtin.ssh").Return(domain.Plugin{
					UID:  "builtin.ssh",
					Name: "SSH",
					Actions: []domain.PluginActionSpec{
						{Action: "terminal", Name: "SSH 终端", BindingUID: "builtin.ssh.host"},
					},
				}, nil)

				repo.EXPECT().ListEnabledBindingsByModelUID(gomock.Any(), "host").Return([]domain.PluginBinding{
					{
						UID:      "builtin.ssh.host",
						PluginID: "builtin.ssh",
						ModelUID: "host",
						Enabled:  true,
					},
				}, nil)

				resolver.EXPECT().LoadResource(gomock.Any(), int64(1), gomock.Any()).Return(domain.Resource{
					ID:       1,
					Name:     "host-01",
					ModelUID: "host",
					Data: map[string]any{
						"ip":       "10.0.0.8",
						"username": "root",
					},
				}, nil)

				resolver.EXPECT().Resolve(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]pluginx.ResolvedInput{
					"target": {
						Name: "target",
						Resources: []pluginx.ResolvedResource{
							{
								ResourceID: 1,
								ModelUID:   "host",
								Fields: map[string]any{
									"ip":       "10.0.0.8",
									"username": "root",
								},
							},
						},
					},
				}, nil)

				return repo, resolver
			},
			req: pluginx.ResolveRequest{
				PluginID:   "builtin.ssh",
				Action:     "terminal",
				ResourceID: 1,
			},
			validate: func(t *testing.T, actCtx pluginx.ActionContext) {
				assert.Equal(t, "builtin.ssh", actCtx.Plugin.UID)
				assert.Equal(t, "terminal", actCtx.Action.Action)
				assert.Equal(t, "10.0.0.8", actCtx.Inputs["target"].Resources[0].Fields["ip"])
				assert.Equal(t, "root", actCtx.Inputs["target"].Resources[0].Fields["username"])
			},
		},
		{
			name: "缺少必需输入时返回友好错误",
			mock: func(ctrl *gomock.Controller) (repository.PluginRepository, IInputResolver) {
				repo := pluginmocks.NewMockPluginRepository(ctrl)
				resolver := pluginmocks.NewMockIInputResolver(ctrl)

				repo.EXPECT().GetPlugin(gomock.Any(), "builtin.ssh").Return(domain.Plugin{
					UID:  "builtin.ssh",
					Name: "SSH",
					Actions: []domain.PluginActionSpec{
						{Action: "terminal", Name: "SSH 终端", BindingUID: "builtin.ssh.host"},
					},
				}, nil)

				repo.EXPECT().ListEnabledBindingsByModelUID(gomock.Any(), "host").Return([]domain.PluginBinding{
					{
						UID:      "builtin.ssh.host",
						PluginID: "builtin.ssh",
						ModelUID: "host",
						Enabled:  true,
					},
				}, nil)

				resolver.EXPECT().LoadResource(gomock.Any(), int64(1), gomock.Any()).Return(domain.Resource{
					ID:       1,
					Name:     "host-01",
					ModelUID: "host",
					Data: map[string]any{
						"username": "root",
					},
				}, nil)

				resolver.EXPECT().Resolve(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, newMissingInputError("target.ip 不能为空"))

				return repo, resolver
			},
			req: pluginx.ResolveRequest{
				PluginID:   "builtin.ssh",
				Action:     "terminal",
				ResourceID: 1,
			},
			wantErr:     true,
			errContains: "target.ip 不能为空",
		},
		{
			name: "插件动作不存在时报错",
			mock: func(ctrl *gomock.Controller) (repository.PluginRepository, IInputResolver) {
				repo := pluginmocks.NewMockPluginRepository(ctrl)
				resolver := pluginmocks.NewMockIInputResolver(ctrl)

				repo.EXPECT().GetPlugin(gomock.Any(), "builtin.ssh").Return(domain.Plugin{
					UID:     "builtin.ssh",
					Name:    "SSH",
					Actions: []domain.PluginActionSpec{},
				}, nil)

				resolver.EXPECT().LoadResource(gomock.Any(), int64(1), gomock.Any()).Return(domain.Resource{
					ID:       1,
					ModelUID: "host",
				}, nil)

				return repo, resolver
			},
			req: pluginx.ResolveRequest{
				PluginID:   "builtin.ssh",
				Action:     "unknown",
				ResourceID: 1,
			},
			wantErr:     true,
			errContains: "插件动作不存在",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, resolver := tc.mock(ctrl)
			svc := &service{
				repo:     repo,
				resolver: resolver,
			}

			actCtx, err := svc.ResolveActionContext(context.Background(), tc.req)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}

			assert.NoError(t, err)
			if tc.validate != nil {
				tc.validate(t, actCtx)
			}
		})
	}
}

func TestGetActionRuntime(t *testing.T) {
	testCases := []struct {
		name        string
		mock        func(ctrl *gomock.Controller) repository.PluginRepository
		pluginID    string
		action      string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, p domain.Plugin, spec pluginx.ActionSpec)
	}{
		{
			name: "成功获取动作运行时定义",
			mock: func(ctrl *gomock.Controller) repository.PluginRepository {
				repo := pluginmocks.NewMockPluginRepository(ctrl)
				repo.EXPECT().GetPlugin(gomock.Any(), "builtin.ssh").Return(domain.Plugin{
					UID:     "builtin.ssh",
					Name:    "SSH",
					Version: "1.0.0",
					Actions: []domain.PluginActionSpec{
						{
							Action: "terminal",
							Name:   "SSH 终端",
							Runtime: &pluginx.ActionRuntimeSpec{
								Layout: "workspace",
								Title:  "Web Shell",
							},
						},
					},
				}, nil)
				return repo
			},
			pluginID: "builtin.ssh",
			action:   "terminal",
			validate: func(t *testing.T, p domain.Plugin, spec pluginx.ActionSpec) {
				assert.Equal(t, "builtin.ssh", p.UID)
				assert.Equal(t, "terminal", spec.Action)
				assert.Equal(t, "workspace", spec.Runtime.Layout)
			},
		},
		{
			name: "参数为空返回错误",
			mock: func(ctrl *gomock.Controller) repository.PluginRepository {
				return pluginmocks.NewMockPluginRepository(ctrl)
			},
			pluginID:    "",
			action:      "terminal",
			wantErr:     true,
			errContains: "不能为空",
		},
		{
			name: "动作不存在返回错误",
			mock: func(ctrl *gomock.Controller) repository.PluginRepository {
				repo := pluginmocks.NewMockPluginRepository(ctrl)
				repo.EXPECT().GetPlugin(gomock.Any(), "builtin.ssh").Return(domain.Plugin{
					UID:     "builtin.ssh",
					Name:    "SSH",
					Actions: []domain.PluginActionSpec{},
				}, nil)
				return repo
			},
			pluginID:    "builtin.ssh",
			action:      "non_exist",
			wantErr:     true,
			errContains: "插件动作不存在",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tc.mock(ctrl)
			svc := &service{repo: repo}

			p, spec, err := svc.GetActionRuntime(context.Background(), tc.pluginID, tc.action)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}

			assert.NoError(t, err)
			if tc.validate != nil {
				tc.validate(t, p, spec)
			}
		})
	}
}

func TestImportDefinitionOnlyPersistsPluginInfo(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.PluginRepository
		def  coreplugin.Definition
	}{
		{
			name: "成功导入插件基本信息且不持久化bindings与schema",
			mock: func(ctrl *gomock.Controller) repository.PluginRepository {
				repo := pluginmocks.NewMockPluginRepository(ctrl)
				repo.EXPECT().UpsertPlugin(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, p domain.Plugin) error {
					assert.Equal(t, "builtin.ssh", p.UID)
					assert.Equal(t, "SSH", p.Name)
					_, hasSchema := p.Meta["schema"]
					assert.False(t, hasSchema)
					return nil
				})
				return repo
			},
			def: coreplugin.Definition{
				Plugin: pluginx.Plugin{
					UID:  "builtin.ssh",
					Name: "SSH",
				},
				Bindings: []pluginx.Binding{
					{
						UID:      "builtin.ssh.host",
						ModelUID: "host",
						Enabled:  true,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tc.mock(ctrl)
			svc := &service{repo: repo}

			err := svc.ImportDefinition(context.Background(), tc.def)
			assert.NoError(t, err)
		})
	}
}

func TestGetDefaultDefinition(t *testing.T) {
	testCases := []struct {
		name        string
		mock        func(ctrl *gomock.Controller) repository.PluginRepository
		pluginID    string
		wantErr     bool
		errContains string
	}{
		{
			name: "空 pluginID 返回错误",
			mock: func(ctrl *gomock.Controller) repository.PluginRepository {
				return pluginmocks.NewMockPluginRepository(ctrl)
			},
			pluginID:    "",
			wantErr:     true,
			errContains: "plugin_id 不能为空",
		},
		{
			name: "运行时不可访问返回友好错误",
			mock: func(ctrl *gomock.Controller) repository.PluginRepository {
				repo := pluginmocks.NewMockPluginRepository(ctrl)
				repo.EXPECT().GetPlugin(gomock.Any(), "builtin.ssh").Return(domain.Plugin{
					UID:  "builtin.ssh",
					Name: "SSH",
					Meta: map[string]any{
						"schema": pluginx.Schema{},
					},
				}, nil)
				return repo
			},
			pluginID:    "builtin.ssh",
			wantErr:     true,
			errContains: "插件默认定义获取失败",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tc.mock(ctrl)
			svc := &service{repo: repo}

			_, err := svc.GetDefaultDefinition(context.Background(), tc.pluginID)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestResolveResult(t *testing.T) {
	testCases := []struct {
		name     string
		ctx      pluginx.ActionContext
		validate func(t *testing.T, res pluginx.ResolveResult)
	}{
		{
			name: "ResolveResult 映射完整字段",
			ctx: pluginx.ActionContext{
				Plugin: pluginx.Plugin{UID: "builtin.ssh", Name: "SSH", Version: "1.0.0"},
				Binding: pluginx.Binding{
					UID:      "builtin.ssh.host",
					ModelUID: "host",
				},
				Action: pluginx.ActionSpec{
					Action:     "terminal",
					Name:       "SSH 终端",
					Permission: "cmdb:ssh:terminal",
					BindingUID: "builtin.ssh.host",
					Runtime: &pluginx.ActionRuntimeSpec{
						Title: "SSH 终端",
					},
					Meta: map[string]any{
						"title": "SSH 终端",
					},
				},
				ResourceID: 42,
			},
			validate: func(t *testing.T, res pluginx.ResolveResult) {
				assert.Equal(t, "host", res.ModelUID)
				assert.Equal(t, "builtin.ssh.host", res.BindingUID)
				assert.Equal(t, "cmdb:ssh:terminal", res.Permission)
				assert.Equal(t, "SSH 终端", res.Meta["title"])
				assert.Equal(t, int64(42), res.ResourceID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := resolveResult(tc.ctx)
			tc.validate(t, res)
		})
	}
}

func TestImportModelRelations(t *testing.T) {
	testCases := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) *relationmocks.MockRelationModelService
		relations []pluginx.ModelRelation
		wantErr   bool
	}{
		{
			name: "已存在且变更则执行更新",
			mock: func(ctrl *gomock.Controller) *relationmocks.MockRelationModelService {
				modelRelations := relationmocks.NewMockRelationModelService(ctrl)
				modelRelations.EXPECT().GetByRelationNames(gomock.Any(), []string{"AuthGateway_default_host"}).
					Return([]domain.ModelRelation{
						{
							ID:              42,
							SourceModelUID:  "AuthGateway",
							TargetModelUID:  "host",
							RelationTypeUID: pluginx.RelationTypeDefault,
							RelationName:    "AuthGateway_default_host",
							Mapping:         pluginx.MappingOneToMany,
						},
					}, nil)

				modelRelations.EXPECT().UpdateModelRelation(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, rel domain.ModelRelation) (int64, error) {
						assert.Equal(t, int64(42), rel.ID)
						assert.Equal(t, pluginx.MappingManyToMany, rel.Mapping)
						return 1, nil
					})

				modelRelations.EXPECT().BatchCreate(gomock.Any(), []domain.ModelRelation{}).Return(nil)
				return modelRelations
			},
			relations: []pluginx.ModelRelation{
				{
					SourceModelUID:  "AuthGateway",
					TargetModelUID:  "host",
					RelationTypeUID: pluginx.RelationTypeDefault,
					Mapping:         pluginx.MappingManyToMany,
				},
			},
		},
		{
			name: "已存在且未变更则跳过更新",
			mock: func(ctrl *gomock.Controller) *relationmocks.MockRelationModelService {
				modelRelations := relationmocks.NewMockRelationModelService(ctrl)
				modelRelations.EXPECT().GetByRelationNames(gomock.Any(), []string{"AuthGateway_default_host"}).
					Return([]domain.ModelRelation{
						{
							ID:              42,
							SourceModelUID:  "AuthGateway",
							TargetModelUID:  "host",
							RelationTypeUID: pluginx.RelationTypeDefault,
							RelationName:    "AuthGateway_default_host",
							Mapping:         pluginx.MappingManyToMany,
						},
					}, nil)

				// 不应触发 UpdateModelRelation
				modelRelations.EXPECT().BatchCreate(gomock.Any(), []domain.ModelRelation{}).Return(nil)
				return modelRelations
			},
			relations: []pluginx.ModelRelation{
				{
					SourceModelUID:  "AuthGateway",
					TargetModelUID:  "host",
					RelationTypeUID: pluginx.RelationTypeDefault,
					Mapping:         pluginx.MappingManyToMany,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			modelRelations := tc.mock(ctrl)
			importer := &schemaImporter{
				modelRelations: modelRelations,
			}

			err := importer.importModelRelations(context.Background(), tc.relations)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
