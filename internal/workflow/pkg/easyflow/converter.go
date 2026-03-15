package easyflow

import (
	"fmt"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/mitchellh/mapstructure"
)

// DefaultConverter 默认转换器实现
type DefaultConverter struct {
	handlers map[string]INodeHandler
}

func NewDefaultConverter() *DefaultConverter {
	return &DefaultConverter{
		handlers: make(map[string]INodeHandler),
	}
}

// NewDefaultConverterWithHandlers 创建已注册所有标准处理器的转换器
func NewDefaultConverterWithHandlers() *DefaultConverter {
	c := NewDefaultConverter()
	c.Register(&StartNodeHandler{})
	c.Register(&EndNodeHandler{})
	c.Register(&UserNodeHandler{})
	c.Register(&ParallelHandler{})
	c.Register(&SelectiveHandler{})
	c.Register(&ConditionHandler{})
	c.Register(&InclusionHandler{})
	c.Register(&AutomationNodeHandler{})
	c.Register(&ChatGroupNodeHandler{})
	return c
}

// Register 注册节点处理器
func (c *DefaultConverter) Register(handler INodeHandler) {
	c.handlers[handler.Type()] = handler
}

// Convert 执行转换流程 (Pipeline)
func (c *DefaultConverter) Convert(wf Workflow) (*model.Process, error) {
	// 注入 context
	ctx, err := c.initContext(wf)
	if err != nil {
		return nil, err
	}

	// 数据转换
	nodes, err := ParseNodes(wf.FlowData.Nodes)
	if err != nil {
		return nil, fmt.Errorf("parse nodes failed: %w", err)
	}

	for _, node := range nodes {
		handler, ok := c.handlers[node.Type]
		if !ok {
			return nil, fmt.Errorf("unsupported node type: %s", node.Type)
		}

		generatedNodes, err := handler.Handle(ctx, node)
		if err != nil {
			return nil, fmt.Errorf("handle node [%s] failed: %w", node.ID, err)
		}
		ctx.OutputNodes = append(ctx.OutputNodes, generatedNodes...)
	}

	// 此处可以扩展如 GraphRewriter, EventInjector 等
	// 目前逻辑简单，直接组装结果

	process := &model.Process{
		ProcessName:  wf.Name,
		Source:       "工单系统",
		RevokeEvents: []string{EventRevoke},
		Nodes:        ctx.OutputNodes,
	}

	return process, nil
}

func (c *DefaultConverter) initContext(wf Workflow) (*Context, error) {
	ctx := &Context{
		Workflow:     wf,
		NodesMap:     make(map[string]Node),
		EdgesMap:     make(map[string][]Edge),
		PrevNodesMap: make(map[string][]string),
		OutputNodes:  []model.Node{},
	}

	edges, err := parseEdges(wf.FlowData.Edges)
	if err != nil {
		return nil, fmt.Errorf("parse edges failed: %w", err)
	}

	nodes, err := ParseNodes(wf.FlowData.Nodes)
	if err != nil {
		return nil, fmt.Errorf("parse nodes failed: %w", err)
	}

	for _, n := range nodes {
		ctx.NodesMap[n.ID] = n
	}

	for _, e := range edges {
		ctx.EdgesMap[e.SourceNodeId] = append(ctx.EdgesMap[e.SourceNodeId], e)
		ctx.PrevNodesMap[e.TargetNodeId] = append(ctx.PrevNodesMap[e.TargetNodeId], e.SourceNodeId)
	}

	return ctx, nil
}

// parseEdges 定义线字段 (搬运自 convert.go)
func parseEdges(raw any) ([]Edge, error) {
	var edges []Edge
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &edges,
		TagName: "json",
	})
	if err != nil {
		return nil, err
	}

	if err = decoder.Decode(raw); err != nil {
		return nil, err
	}

	return edges, nil
}
