package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	pluginv1 "github.com/Duke1616/ecmdb/api/proto/gen/ecmdb/plugin/v1"
	"github.com/Duke1616/ecmdb/internal/service/plugin"
	mongoxplugin "github.com/Duke1616/ecmdb/pkg/mongox/plugin"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
)

type Server struct {
	pluginv1.UnimplementedPluginRuntimeServiceServer
	svc plugin.Service
}

func NewServer(svc plugin.Service) *Server {
	return &Server{svc: svc}
}

// ResolveActionContext 暴露给插件端通过 gRPC 远程调用，解析具体的上下文
func (s *Server) ResolveActionContext(ctx context.Context, req *pluginv1.ResolveActionContextRequest) (*pluginv1.ResolveActionContextResponse, error) {
	resolveReq := pluginx.ResolveRequest{
		PluginID:   req.PluginId,
		Action:     req.Action,
		ResourceID: req.ResourceId,
	}

	actionCtx, err := s.svc.ResolveActionContext(ctx, resolveReq)
	if err != nil {
		return nil, err
	}

	// 序列化为 JSON 字节响应
	actionCtxJSON, err := json.Marshal(actionCtx)
	if err != nil {
		return nil, err
	}

	return &pluginv1.ResolveActionContextResponse{
		ActionContextJson: actionCtxJSON,
	}, nil
}

// RegisterPlugin 插件端在启动后通过此 gRPC 接口将自己注册给主站，主站反向拉取自描述信息并自动配置导入
func (s *Server) RegisterPlugin(ctx context.Context, req *pluginv1.RegisterPluginRequest) (*pluginv1.RegisterPluginResponse, error) {
	if req.Upstream == "" {
		return nil, fmt.Errorf("upstream 地址不能为空")
	}

	def, err := pluginx.FetchDefinition(ctx, req.Upstream)
	if err != nil {
		return nil, fmt.Errorf("读取插件自描述 Definition 失败: %w", err)
	}

	// 保证落库的 upstream 地址采用插件注册时传入的地址
	spec, ok := def.Plugin.Runtime()
	if !ok {
		spec = pluginx.RuntimeSpec{
			Mode: pluginx.RuntimeModeExternalService,
		}
	}
	spec.Upstream = req.Upstream
	def.Plugin.SetRuntime(spec)

	// 内置插件属于全局共享数据（tenant_id=0），仅此注册场景需要绑过租户隔离，
	// 确保 Upsert 时能命中已有记录走 Update 而非 Insert（否则会触发主键冲突）
	registerCtx := mongoxplugin.IgnoreTenantContext(ctx)

	// 3. 调用主站内置服务直接进行注册落库
	if err = s.svc.ImportDefinition(registerCtx, def); err != nil {
		return nil, fmt.Errorf("主站导入插件自描述 Definition 失败: %w", err)
	}

	return &pluginv1.RegisterPluginResponse{
		Success: true,
	}, nil
}
