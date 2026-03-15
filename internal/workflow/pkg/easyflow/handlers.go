package easyflow

import (
	"github.com/Bunny3th/easy-workflow/workflow/model"
)

// StartNodeHandler 处理开始节点
type StartNodeHandler struct{}

func (h *StartNodeHandler) Type() string { return NodeTypeStart }
func (h *StartNodeHandler) Handle(ctx *Context, node Node) ([]model.Node, error) {
	nodeName := DefaultNameStart
	property, _ := ToNodeProperty[StartProperty](node)
	if property.Name != "" {
		nodeName = property.Name
	}
	n := model.Node{
		NodeID:        node.ID,
		NodeName:      nodeName,
		NodeType:      0,
		UserIDs:       []string{UserStarter},
		NodeEndEvents: []string{EventStart},
	}
	return []model.Node{n}, nil
}

// EndNodeHandler 处理结束节点
type EndNodeHandler struct{}

func (h *EndNodeHandler) Type() string { return NodeTypeEnd }
func (h *EndNodeHandler) Handle(ctx *Context, node Node) ([]model.Node, error) {
	nodeName := DefaultNameEnd
	property, _ := ToNodeProperty[EndProperty](node)
	if property.Name != "" {
		nodeName = property.Name
	}
	n := model.Node{
		NodeID:          node.ID,
		NodeName:        nodeName,
		NodeType:        3,
		PrevNodeIDs:     ctx.GetPrevNodeIDs(node.ID),
		NodeStartEvents: []string{EventNotify},
	}
	return []model.Node{n}, nil
}

// UserNodeHandler 处理用户/审批节点
type UserNodeHandler struct{}

func (h *UserNodeHandler) Type() string { return NodeTypeUser }
func (h *UserNodeHandler) Handle(ctx *Context, node Node) ([]model.Node, error) {
	property, _ := ToNodeProperty[UserProperty](node)
	nodeName := DefaultNameUser
	if property.Name != "" {
		nodeName = property.Name
	}

	startEvents := []string{EventNotify}
	if property.IsCC {
		startEvents = []string{EventCarbonCopy}
	}

	userIDs := property.Approved
	if userIDs == nil {
		userIDs = []string{}
	}

	n := model.Node{
		NodeID:          node.ID,
		NodeName:        nodeName,
		NodeType:        1,
		PrevNodeIDs:     findAndProxySrcNodes(ctx, node),
		UserIDs:         userIDs,
		IsCosigned:      0,
		NodeStartEvents: startEvents,
	}

	if property.IsCosigned {
		n.IsCosigned = 1
	}

	n.TaskFinishEvents = append(getPassEvents(ctx, node.ID), getRejectEvents(ctx, node.ID)...)

	return []model.Node{n}, nil
}

// Gateway Handlers

type ParallelHandler struct{}

func (h *ParallelHandler) Type() string { return NodeTypeParallel }
func (h *ParallelHandler) Handle(ctx *Context, node Node) ([]model.Node, error) {
	edgesDst := ctx.GetTargetEdges(node.ID)
	inevitableNodes := make([]string, 0, len(edgesDst))
	for _, edge := range edgesDst {
		inevitableNodes = append(inevitableNodes, ctx.GetOrGenerateProxyID(node.ID, edge.TargetNodeId))
	}
	gw := model.HybridGateway{
		Conditions:         nil,
		InevitableNodes:    inevitableNodes,
		WaitForAllPrevNode: 1,
	}
	n := model.Node{
		NodeID:      node.ID,
		NodeName:    DefaultNameParallel,
		NodeType:    2,
		GWConfig:    gw,
		PrevNodeIDs: findAndProxySrcNodesForGateway(ctx, node),
	}
	return []model.Node{n}, nil
}

type InclusionHandler struct{}

func (h *InclusionHandler) Type() string { return NodeTypeInclusion }
func (h *InclusionHandler) Handle(ctx *Context, node Node) ([]model.Node, error) {
	edgesDst := ctx.GetTargetEdges(node.ID)
	inevitableNodes := make([]string, 0, len(edgesDst))
	for _, edge := range edgesDst {
		inevitableNodes = append(inevitableNodes, ctx.GetOrGenerateProxyID(node.ID, edge.TargetNodeId))
	}
	gw := model.HybridGateway{
		Conditions:         nil,
		InevitableNodes:    inevitableNodes,
		WaitForAllPrevNode: 0,
	}
	n := model.Node{
		NodeID:      node.ID,
		NodeName:    DefaultNameInclusion,
		NodeType:    2,
		GWConfig:    gw,
		PrevNodeIDs: findAndProxySrcNodesForGateway(ctx, node),
	}
	return []model.Node{n}, nil
}

type SelectiveHandler struct{}

func (h *SelectiveHandler) Type() string { return NodeTypeSelective }
func (h *SelectiveHandler) Handle(ctx *Context, node Node) ([]model.Node, error) {
	edgesDst := ctx.GetTargetEdges(node.ID)
	inevitableNodes := make([]string, 0, len(edgesDst))
	for _, edge := range edgesDst {
		inevitableNodes = append(inevitableNodes, ctx.GetOrGenerateProxyID(node.ID, edge.TargetNodeId))
	}
	gw := model.HybridGateway{
		Conditions:         nil,
		InevitableNodes:    inevitableNodes,
		WaitForAllPrevNode: 1,
	}
	n := model.Node{
		NodeID:          node.ID,
		NodeName:        DefaultNameSelective,
		NodeType:        2,
		GWConfig:        gw,
		PrevNodeIDs:     findAndProxySrcNodesForGateway(ctx, node),
		NodeStartEvents: []string{EventSelectiveGatewaySplit},
	}
	return []model.Node{n}, nil
}

type ConditionHandler struct{}

func (h *ConditionHandler) Type() string { return NodeTypeCondition }
func (h *ConditionHandler) Handle(ctx *Context, node Node) ([]model.Node, error) {
	property, _ := ToNodeProperty[ConditionProperty](node)
	edgesDst := ctx.GetTargetEdges(node.ID)
	conditions := make([]model.Condition, 0, len(edgesDst))
	for _, edge := range edgesDst {
		edgeProp, _ := ToEdgeProperty(edge)
		expr := edgeProp.Expression
		if expr == "" {
			expr = "1 = 1"
		}
		conditions = append(conditions, model.Condition{
			Expression: expr,
			NodeID:     ctx.GetOrGenerateProxyID(node.ID, edge.TargetNodeId),
		})
	}
	gw := model.HybridGateway{
		Conditions:         conditions,
		InevitableNodes:    []string{},
		WaitForAllPrevNode: 3,
	}
	n := model.Node{
		NodeID:      node.ID,
		NodeName:    property.Name,
		NodeType:    2,
		GWConfig:    gw,
		PrevNodeIDs: ctx.GetPrevNodeIDs(node.ID),
	}
	return []model.Node{n}, nil
}

type AutomationNodeHandler struct{}

func (h *AutomationNodeHandler) Type() string { return NodeTypeAuto }
func (h *AutomationNodeHandler) Handle(ctx *Context, node Node) ([]model.Node, error) {
	property, _ := ToNodeProperty[AutomationProperty](node)
	n := model.Node{
		NodeID:          node.ID,
		NodeName:        property.Name,
		NodeType:        1,
		PrevNodeIDs:     ctx.GetPrevNodeIDs(node.ID),
		UserIDs:         []string{AutomationApproval},
		NodeStartEvents: []string{EventAutomation},
		NodeEndEvents:   []string{EventNotify},
	}
	return []model.Node{n}, nil
}

type ChatGroupNodeHandler struct{}

func (h *ChatGroupNodeHandler) Type() string { return NodeTypeChat }
func (h *ChatGroupNodeHandler) Handle(ctx *Context, node Node) ([]model.Node, error) {
	property, _ := ToNodeProperty[ChatGroupProperty](node)
	n := model.Node{
		NodeID:          node.ID,
		NodeName:        property.Name,
		NodeType:        1,
		PrevNodeIDs:     ctx.GetPrevNodeIDs(node.ID),
		UserIDs:         []string{ChatGroupApproval},
		NodeStartEvents: []string{EventChatGroup},
	}
	return []model.Node{n}, nil
}

// 辅助函数

func isGateway(t string) bool {
	return t == NodeTypeParallel || t == NodeTypeInclusion || t == NodeTypeSelective
}

func isConditionGateway(t string) bool {
	return t == NodeTypeCondition
}

func findAndProxySrcNodes(ctx *Context, node Node) []string {
	prevIDs := ctx.GetPrevNodeIDs(node.ID)
	result := make([]string, 0, len(prevIDs))
	for _, pid := range prevIDs {
		// 用户节点不需要代理，直接连接
		result = append(result, pid)
	}
	return result
}

func findAndProxySrcNodesForGateway(ctx *Context, node Node) []string {
	prevIDs := ctx.GetPrevNodeIDs(node.ID)
	result := make([]string, 0, len(prevIDs))
	for _, pid := range prevIDs {
		prevNode := ctx.GetNodeInfo(pid)
		// 只有当前置节点是条件网关时，才需要代理
		if isConditionGateway(prevNode.Type) {
			result = append(result, ctx.GetOrGenerateProxyForGateway(pid, node.ID))
		} else {
			result = append(result, pid)
		}
	}
	return result
}

func getPassEvents(ctx *Context, nodeID string) []string {
	var events []string
	edges := ctx.GetTargetEdges(nodeID)
	existEvent := make(map[string]struct{})

	for _, edge := range edges {
		info := ctx.GetNodeInfo(edge.TargetNodeId)
		switch info.Type {
		case NodeTypeParallel:
			addEvent(&events, existEvent, EventTaskParallelNodePass)
		case NodeTypeInclusion:
			addEvent(&events, existEvent, EventTaskInclusionNodePass)
		}
	}
	return events
}

func getRejectEvents(ctx *Context, nodeID string) []string {
	var events []string
	existEvent := make(map[string]struct{})
	parents := getParents(ctx, nodeID)

	for _, p := range parents {
		grandParents := getParents(ctx, p.ID)
		if p.Type == NodeTypeCondition && hasType(grandParents, NodeTypeParallel, NodeTypeInclusion, NodeTypeSelective) {
			addEvent(&events, existEvent, EventConcurrentRejectCleanup)
		}
		if p.Type == NodeTypeCondition && hasType(grandParents, NodeTypeInclusion) {
			addEvent(&events, existEvent, EventInclusionPassCleanup)
		}
		if isGateway(p.Type) && hasType(grandParents, NodeTypeCondition) {
			addEvent(&events, existEvent, EventGatewayConditionReject)
		}
		if isGateway(p.Type) && hasProxyInGateway(ctx, p.ID) {
			addEvent(&events, existEvent, EventUserNodeRejectProxyCleanup)
		}
	}
	return events
}

type nodeWithInfo struct {
	ID   string
	Type string
}

func getParents(ctx *Context, nodeID string) []nodeWithInfo {
	prevIDs := ctx.GetPrevNodeIDs(nodeID)
	res := make([]nodeWithInfo, 0, len(prevIDs))
	for _, pid := range prevIDs {
		n := ctx.GetNodeInfo(pid)
		res = append(res, nodeWithInfo{ID: n.ID, Type: n.Type})
	}
	return res
}

func hasType(nodes []nodeWithInfo, types ...string) bool {
	typeSet := make(map[string]struct{})
	for _, t := range types {
		typeSet[t] = struct{}{}
	}
	for _, n := range nodes {
		if _, ok := typeSet[n.Type]; ok {
			return true
		}
	}
	return false
}

func addEvent(events *[]string, exist map[string]struct{}, event string) {
	if _, ok := exist[event]; !ok {
		*events = append(*events, event)
		exist[event] = struct{}{}
	}
}

func hasProxyInGateway(ctx *Context, gatewayId string) bool {
	edges := ctx.GetTargetEdges(gatewayId)
	for _, edge := range edges {
		if checkBranchHasProxy(ctx, edge.TargetNodeId, make(map[string]bool)) {
			return true
		}
	}
	return false
}

func checkBranchHasProxy(ctx *Context, nodeID string, visited map[string]bool) bool {
	if visited[nodeID] {
		return false
	}
	visited[nodeID] = true
	node := ctx.GetNodeInfo(nodeID)
	if node.Type == NodeTypeCondition {
		return true
	}
	edges := ctx.GetTargetEdges(nodeID)
	for _, edge := range edges {
		if checkBranchHasProxy(ctx, edge.TargetNodeId, visited) {
			return true
		}
	}
	return false
}
