package plugin

import (
	"context"
	"fmt"
	"strings"

	"github.com/Duke1616/ecmdb/pkg/plugin/codec"
	"github.com/Duke1616/ecmdb/pkg/plugin/graph"
	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
)

// TargetBuilder 代表针对特定目标资产（及其关联拓扑图谱）的动作构建器
type TargetBuilder[T any] struct {
	registry   *Registry
	modelUID   string
	bindingUID string
}

// Target 在插件注册表上声明受控的目标资产 T 与根模型 modelUID
// 自动根据泛型 T 递归推导派生 CMDB 模型定义、属性与关联关系图谱，构建绑定
func Target[T any](r *Registry, modelUID string) *TargetBuilder[T] {
	if r.err != nil {
		return &TargetBuilder[T]{registry: r}
	}
	modelUID = strings.TrimSpace(modelUID)
	bindingUID := fmt.Sprintf("%s.%s", r.plugin.UID, modelUID)

	// 1. 单次深度内省泛型 T，同时推导出 CMDB Schema 与 ResourceSpec 图谱
	meta, err := codec.InspectTarget[T]("target", modelUID)
	if err != nil {
		r.err = err
		return &TargetBuilder[T]{registry: r}
	}
	r.schema.Merge(meta.Schema)

	// 2. 构建并编译 Binding Graph
	g, err := graph.GraphFromBindingSpecs(modelUID, []types.ResourceSpec{meta.Spec})
	if err != nil {
		r.err = err
		return &TargetBuilder[T]{registry: r}
	}

	binding := types.Binding{
		UID:      bindingUID,
		PluginID: r.plugin.UID,
		ModelUID: modelUID,
		Enabled:  true,
		Graph:    g,
	}
	if _, err = graph.CompileBindingGraph(binding.Graph); err != nil {
		r.err = err
		return &TargetBuilder[T]{registry: r}
	}

	if !lo.SomeBy(r.bindings, func(b types.Binding) bool { return b.UID == binding.UID }) {
		r.bindings = append(r.bindings, binding)
	}

	return &TargetBuilder[T]{
		registry:   r,
		modelUID:   modelUID,
		bindingUID: bindingUID,
	}
}

// Model 显式声明或覆盖当前受控目标资产模型的展示名称与所属分组（优先级最高）
// 适用于单模型插件或需要覆盖结构体内自描述信息的场景
func (b *TargetBuilder[T]) Model(name, group string) *TargetBuilder[T] {
	if b.registry.err == nil {
		b.registry.schema.UpdateModelMeta(b.modelUID, name, group)
	}
	return b
}

// Workspace 围绕当前目标资产挂载沉浸式工作台看板动作
// 侧边栏资源展示与数据图谱天然对齐该目标资产，消除重复声明
func (b *TargetBuilder[T]) Workspace(action string, title string, opts ...ActionOption) *TargetBuilder[T] {
	spec := types.ActionSpec{
		Action:     action,
		Name:       title,
		BindingUID: b.bindingUID,
		Placement:  types.PlacementResourceDetailActions,
	}

	runtime := newDefaultWorkspaceRuntime(title, b.modelUID)
	spec.Runtime = &runtime

	for _, opt := range opts {
		opt(&spec)
	}

	b.registry.plugin.Actions = append(b.registry.plugin.Actions, spec)
	return b
}

// Action 围绕当前目标资产挂载通用操作动作（如详情页按钮、快捷运维等），天然继承该资产的数据绑定
func (b *TargetBuilder[T]) Action(action string, name string, opts ...ActionOption) *TargetBuilder[T] {
	spec := types.ActionSpec{
		Action:     action,
		Name:       name,
		BindingUID: b.bindingUID,
		Placement:  types.PlacementResourceDetailActions,
	}
	for _, opt := range opts {
		opt(&spec)
	}
	b.registry.plugin.Actions = append(b.registry.plugin.Actions, spec)
	return b
}

// Registry 返回根注册表实例
func (b *TargetBuilder[T]) Registry() *Registry {
	return b.registry
}

// Definition 编译并生成最终的插件自描述元数据
func (b *TargetBuilder[T]) Definition() (Definition, error) {
	return b.registry.Definition()
}

// MustDefinition 编译并生成最终的插件自描述元数据，若出错则 panic（常用于单元测试）
func (b *TargetBuilder[T]) MustDefinition() Definition {
	return b.registry.MustDefinition()
}

// ResolveActionRoot 从 ECMDB 控制面拉取指定动作的解密上下文，并一键解码为强类型结构 T
func ResolveActionRoot[T any](ctx context.Context, resolver ContextResolver, pluginID, action string, resourceID int64) (T, error) {
	var zero T
	if resolver == nil {
		return zero, nil
	}
	actionCtx, err := resolver.ResolveActionContext(ctx, types.ResolveRequest{
		PluginID:   pluginID,
		Action:     action,
		ResourceID: resourceID,
	})
	if err != nil {
		return zero, err
	}
	return codec.InputRootOne[T](actionCtx)
}
