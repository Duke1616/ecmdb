package plugin

import (
	"context"
	"errors"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/internal/repository"
	attribute "github.com/Duke1616/ecmdb/internal/service/attribute"
	model "github.com/Duke1616/ecmdb/internal/service/model"
	relation "github.com/Duke1616/ecmdb/internal/service/relation"
	resource "github.com/Duke1616/ecmdb/internal/service/resource"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
	"github.com/samber/lo"
)

type Service interface {
	// RegisterBuiltinPlugins 仅注册引导注册系统内置插件元数据（包含 input_specs 元数据约束表）。
	RegisterBuiltinPlugins(ctx context.Context) error

	// SyncDefaultSchema 供租户在前端按需一键同步激活内置模型的 Model 和默认 Binding。
	SyncDefaultSchema(ctx context.Context, pluginID string) error

	// RegisterDefinition 保存 SDK 注册出来的插件定义包。
	RegisterDefinition(ctx context.Context, def pluginx.Definition) error

	// UpsertPlugin 创建或更新插件定义。
	UpsertPlugin(ctx context.Context, p pluginx.Plugin) error

	// UpsertBinding 创建或更新插件和资源模型之间的输入树。
	UpsertBinding(ctx context.Context, b pluginx.Binding) error

	// ListPlugins 查询插件目录。
	ListPlugins(ctx context.Context) ([]PluginListItem, error)

	// GetPluginDetail 查询插件详情。
	GetPluginDetail(ctx context.Context, uid string) (PluginDetail, error)

	// ListEnums 查询插件管理所需枚举。
	ListEnums(ctx context.Context) (PluginManagementEnums, error)

	// DeletePlugin 删除插件。
	DeletePlugin(ctx context.Context, uid string) error

	// ListResourceActions 查询指定资源可以使用的插件动作。
	ListResourceActions(ctx context.Context, resourceID int64) ([]pluginx.ResourceAction, error)

	// ListModelActions 查询指定模型配置的插件动作。
	ListModelActions(ctx context.Context, modelUID string) ([]pluginx.ResourceAction, error)

	// ResolveAction 解析插件动作需要的 UI 和输入数据。
	ResolveAction(ctx context.Context, req pluginx.ResolveRequest) (pluginx.ResolveResult, error)

	// ResolveActionContext 解析插件动作运行时上下文，供内置后端能力直接复用。
	ResolveActionContext(ctx context.Context, req pluginx.ResolveRequest) (pluginx.ActionContext, error)
}

type service struct {
	repo           repository.PluginRepository
	resolver       *inputResolver
	models         model.Service
	modelGroups    model.MGService
	attributes     attribute.Service
	relationTypes  relation.RelationTypeService
	modelRelations relation.RelationModelService
}

type actionTarget struct {
	resource domain.Resource
	binding  domain.PluginBinding
	plugin   domain.Plugin
	action   domain.PluginActionSpec
}

func NewService(
	repo repository.PluginRepository,
	resourceSvc resource.Service,
	relationSvc relation.RelationResourceService,
	modelSvc model.Service,
	modelGroupSvc model.MGService,
	attributeSvc attribute.Service,
	relationTypeSvc relation.RelationTypeService,
	relationModelSvc relation.RelationModelService,
) Service {
	return &service{
		repo:           repo,
		models:         modelSvc,
		modelGroups:    modelGroupSvc,
		attributes:     attributeSvc,
		relationTypes:  relationTypeSvc,
		modelRelations: relationModelSvc,
		resolver: newInputResolver(
			resourceSvc,
			relationSvc,
		),
	}
}

func (s *service) RegisterDefinition(ctx context.Context, def pluginx.Definition) error {
	return s.importDefinition(ctx, def)
}

func (s *service) UpsertPlugin(ctx context.Context, p pluginx.Plugin) error {
	if err := validatePlugin(p); err != nil {
		return err
	}
	return s.repo.UpsertPlugin(ctx, p)
}

func (s *service) UpsertBinding(ctx context.Context, b pluginx.Binding) error {
	if err := validateBinding(b); err != nil {
		return err
	}
	return s.repo.UpsertBinding(ctx, b)
}

func (s *service) ResolveAction(ctx context.Context, req pluginx.ResolveRequest) (pluginx.ResolveResult, error) {
	actionCtx, err := s.ResolveActionContext(ctx, req)
	if err != nil {
		return pluginx.ResolveResult{}, err
	}
	return resolveResult(actionCtx), nil
}

func (s *service) ResolveActionContext(ctx context.Context, req pluginx.ResolveRequest) (pluginx.ActionContext, error) {
	if err := validateResolveRequest(req); err != nil {
		return pluginx.ActionContext{}, err
	}

	target, err := s.loadActionTarget(ctx, req)
	if err != nil {
		return pluginx.ActionContext{}, err
	}

	inputs, err := s.resolver.resolve(ctx, target.resource, target.binding.Specs)
	if err != nil {
		if errors.Is(err, errRequiredInputMissing) {
			return pluginx.ActionContext{}, errs.ValidationError.WithMsg(
				fmt.Sprintf("插件动作缺少必需输入: %s", missingInputMessage(err)),
			)
		}
		return pluginx.ActionContext{}, err
	}

	return pluginx.ActionContext{
		Plugin:     target.plugin,
		Binding:    target.binding,
		Action:     target.action,
		ResourceID: req.ResourceID,
		Inputs:     inputs,
		Params:     req.Params,
	}, nil
}

func (s *service) loadActionTarget(ctx context.Context, req pluginx.ResolveRequest) (actionTarget, error) {
	primary, err := s.resolver.loadResource(ctx, req.ResourceID, nil)
	if err != nil {
		return actionTarget{}, err
	}

	binding, err := s.findBinding(ctx, primary.ModelUID, req.PluginID)
	if err != nil {
		return actionTarget{}, err
	}

	plugin, err := s.repo.GetPlugin(ctx, binding.PluginID)
	if err != nil {
		return actionTarget{}, err
	}

	action, ok := findAction(plugin.Actions, req.Action)
	if !ok {
		return actionTarget{}, fmt.Errorf("插件动作不存在: %s", req.Action)
	}

	return actionTarget{
		resource: primary,
		binding:  binding,
		plugin:   plugin,
		action:   action,
	}, nil
}

func (s *service) loadPlugin(ctx context.Context, pluginID string) (domain.Plugin, error) {
	return s.repo.GetPlugin(ctx, pluginID)
}

func (s *service) findBinding(ctx context.Context, modelUID string, pluginID string) (domain.PluginBinding, error) {
	bindings, err := s.repo.ListEnabledBindingsByModelUID(ctx, modelUID)
	if err != nil {
		return domain.PluginBinding{}, err
	}
	binding, ok := lo.Find(bindings, func(binding domain.PluginBinding) bool {
		return binding.PluginID == pluginID
	})
	if !ok {
		return domain.PluginBinding{}, fmt.Errorf("插件绑定不存在")
	}
	return binding, nil
}

func resolveResult(actionCtx pluginx.ActionContext) pluginx.ResolveResult {
	return pluginx.ResolveResult{
		UI:         actionCtx.Action.UI,
		PluginID:   actionCtx.Plugin.UID,
		PluginName: actionCtx.Plugin.Name,
		Action:     actionCtx.Action.Action,
		ResourceID: actionCtx.ResourceID,
		Inputs:     actionCtx.Inputs,
		Params:     actionCtx.Params,
		Meta:       actionCtx.Action.Meta,
	}
}
