package easyflow

import (
	"encoding/json"
	"sync"

	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/ecodeclub/ekit/slice"
)

const (
	AutomationApproval = "automation"
	SysAutoUser        = "sys_auto"
	SysProxyNodeName   = "系统代理流转"
)

type logicFlow struct {
	Workflow Workflow
	Edges    []Edge
	Nodes    []Node

	// 后端存储结构体
	NodeList []model.Node
	mu       sync.Mutex
}

func NewLogicFlowToEngineConvert() ProcessEngineConvert {
	return &logicFlow{
		mu: sync.Mutex{},
	}
}

func (l *logicFlow) GetAutomationProperty(workflow Workflow, nodeId string) (AutomationProperty, error) {
	nodesJSON, err := json.Marshal(workflow.FlowData.Nodes)
	if err != nil {
		return AutomationProperty{}, err
	}

	var nodes []Node
	err = json.Unmarshal(nodesJSON, &nodes)
	if err != nil {
		return AutomationProperty{}, err
	}

	property := AutomationProperty{}
	for _, node := range nodes {
		if node.ID == nodeId {
			property, _ = ToNodeProperty[AutomationProperty](node)
		}
	}

	return property, err
}

func (l *logicFlow) Edge(workflow Workflow, tasks []model.Task) ([]string, error) {
	return nil, nil
}

func (l *logicFlow) Deploy(workflow Workflow) (int, error) {
	// 加锁
	l.mu.Lock()
	defer l.mu.Unlock()

	// 赋值数据
	l.Workflow = workflow

	// 清空 NodeList
	l.NodeList = []model.Node{}

	if err := l.toEdges(); err != nil {
		return 0, err
	}
	if err := l.toNodes(); err != nil {
		return 0, err
	}

	for _, node := range l.Nodes {
		switch node.Type {
		case "start":
			l.Start(node)
		case "end":
			l.End(node)
		case "user":
			l.User(node)
		case "condition":
			l.Condition(node)
		case "parallel":
			l.Parallel(node)
		case "inclusion":
			l.Inclusion(node)
		case "automation":
			l.Automation(node)
		}
	}

	// 发布流程
	process := model.Process{ProcessName: l.Workflow.Name, Source: "工单系统", RevokeEvents: []string{"EventRevoke"}, Nodes: l.NodeList}

	// 列表重新为空
	l.NodeList = nil

	j, err := engine.JSONMarshal(process, false)
	if err != nil {
		return 0, err
	}

	return engine.ProcessSave(string(j), l.Workflow.Owner)
}

func (l *logicFlow) Start(node Node) {
	NodeName := "Start"
	property, _ := ToNodeProperty[StartProperty](node)
	if property.Name != "" {
		NodeName = property.Name
	}
	n := model.Node{NodeID: node.ID, NodeName: NodeName,
		NodeType: 0, UserIDs: []string{"$starter"},
		NodeEndEvents: []string{"EventStart"},
	}

	l.NodeList = append(l.NodeList, n)
}

func (l *logicFlow) End(node Node) {
	NodeName := "End"
	property, _ := ToNodeProperty[EndProperty](node)
	if property.Name != "" {
		NodeName = property.Name
	}
	n := model.Node{NodeID: node.ID, NodeName: NodeName,
		NodeType: 3, PrevNodeIDs: l.FindPrevNodeIDs(node.ID),
		NodeStartEvents: []string{"EventNotify"},
	}
	l.NodeList = append(l.NodeList, n)
}

func (l *logicFlow) Automation(node Node) {
	NodeName := "自动化节点"
	property, _ := ToNodeProperty[AutomationProperty](node)
	if property.Name != "" {
		NodeName = property.Name
	}

	n := model.Node{NodeID: node.ID, NodeName: NodeName,
		NodeType: 1, PrevNodeIDs: l.FindPrevNodeIDs(node.ID),
		UserIDs:          []string{AutomationApproval},
		NodeStartEvents:  []string{"EventAutomation"},
		NodeEndEvents:    []string{"EventNotify"},
		TaskFinishEvents: l.getPassEvents(node.ID),
	}
	l.NodeList = append(l.NodeList, n)
}

func (l *logicFlow) Parallel(node Node) {
	// 查看下级 node 节点 id
	edgesDst := l.FindTargetNodeId(node.ID)
	InevitableNodes := slice.Map(edgesDst, func(idx int, src Edge) string {
		return l.getProxyTargetNodeId(node.ID, src.TargetNodeId)
	})
	gwParallel := model.HybridGateway{Conditions: nil, InevitableNodes: InevitableNodes, WaitForAllPrevNode: 1}

	// 查看上级 node 节点 id
	preNodeIds := l.findAndProxySrcNodes(node)

	n := model.Node{NodeID: node.ID, NodeName: "并行网关",
		NodeType: 2, GWConfig: gwParallel,
		PrevNodeIDs: preNodeIds,
	}

	l.NodeList = append(l.NodeList, n)
}

func (l *logicFlow) Inclusion(node Node) {
	// 查看下级 node 节点 id
	edgesDst := l.FindTargetNodeId(node.ID)
	InevitableNodes := slice.Map(edgesDst, func(idx int, src Edge) string {
		return l.getProxyTargetNodeId(node.ID, src.TargetNodeId)
	})

	gwParallel := model.HybridGateway{Conditions: nil, InevitableNodes: InevitableNodes, WaitForAllPrevNode: 0}
	preNodeIds := l.findAndProxySrcNodes(node)

	n := model.Node{NodeID: node.ID, NodeName: "包容网关",
		NodeType: 2, GWConfig: gwParallel,
		PrevNodeIDs: preNodeIds,
	}

	l.NodeList = append(l.NodeList, n)
}

func (l *logicFlow) Condition(node Node) {
	// 获取所有判断条件的连接线
	edgesDst := l.FindTargetNodeId(node.ID)

	// 组合 conditions 条件
	conditions := slice.Map(edgesDst, func(idx int, src Edge) model.Condition {
		property, _ := ToEdgeProperty(src)

		// 如果没有设置表达式，默认设置为 1 = 1, 自动通过
		expression := property.Expression
		if expression == "" {
			expression = "1 = 1"
		}

		return model.Condition{
			Expression: expression,
			NodeID:     l.getProxyTargetNodeId(node.ID, src.TargetNodeId),
		}
	})

	// 拼接网关
	// 如果存在两个连续的Condition网关, 会导致 easy-workflow 内部处理出现问题， WaitForAllPrevNode = 3
	GwCondition := model.HybridGateway{Conditions: conditions, InevitableNodes: []string{}, WaitForAllPrevNode: 3}

	// node 节点录入
	property, _ := ToNodeProperty[ConditionProperty](node)
	n := model.Node{NodeID: node.ID, NodeName: property.Name,
		NodeType: 2, GWConfig: GwCondition,
		PrevNodeIDs: l.FindPrevNodeIDs(node.ID),
	}

	l.NodeList = append(l.NodeList, n)
}

func (l *logicFlow) User(node Node) {
	// node 节点录入
	property, _ := ToNodeProperty[UserProperty](node)

	// 判断是否为会签节点
	IsCosigned := 0
	if property.IsCosigned == true {
		IsCosigned = 1
	}

	// 录入数据
	n := model.Node{NodeID: node.ID, NodeName: property.Name,
		NodeType: 1, UserIDs: property.Approved,
		PrevNodeIDs:      l.FindPrevNodeIDs(node.ID),
		TaskFinishEvents: append(l.getPassEvents(node.ID), l.getRejectEvents(node.ID)...),
		NodeStartEvents:  []string{"EventNotify"},
		IsCosigned:       IsCosigned,
	}

	l.NodeList = append(l.NodeList, n)
}

func (l *logicFlow) getPassEvents(nodeId string) []string {
	var events []string
	edges := l.FindTargetNodeId(nodeId)
	existEvent := make(map[string]struct{})

	for _, edge := range edges {
		info := l.GetNodeInfo(edge.TargetNodeId)
		switch info.Type {
		case "parallel":
			if _, ok := existEvent["EventTaskParallelNodePass"]; !ok {
				events = append(events, "EventTaskParallelNodePass")
				existEvent["EventTaskParallelNodePass"] = struct{}{}
			}
		case "inclusion":
			if _, ok := existEvent["EventTaskInclusionNodePass"]; !ok {
				events = append(events, "EventTaskInclusionNodePass")
				existEvent["EventTaskInclusionNodePass"] = struct{}{}
			}
		}
	}
	return events
}

type nodeWithInfo struct {
	ID   string
	Type string
}

// 获取某个节点的直接上级节点（带类型）
func (l *logicFlow) parents(nodeId string) []nodeWithInfo {
	edges := l.FindSourceNodeId(nodeId)
	res := make([]nodeWithInfo, 0, len(edges))

	for _, e := range edges {
		n := l.GetNodeInfo(e.SourceNodeId)
		res = append(res, nodeWithInfo{
			ID:   n.ID,
			Type: n.Type,
		})
	}

	return res
}

// 判断节点列表中是否存在某种类型
func hasType(nodes []nodeWithInfo, types ...string) bool {
	typeSet := make(map[string]struct{}, len(types))
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

// 添加事件（自动去重）
func addEvent(events *[]string, exist map[string]struct{}, event string) {
	if _, ok := exist[event]; ok {
		return
	}
	*events = append(*events, event)
	exist[event] = struct{}{}
}

func isGateway(t string) bool {
	return t == "parallel" || t == "inclusion"
}

func (l *logicFlow) getRejectEvents(nodeId string) []string {
	events := make([]string, 0, 2)
	existEvent := make(map[string]struct{})

	// Me 的直接上级
	parents := l.parents(nodeId)

	for _, p := range parents {
		// Me 的上级的上级
		grandParents := l.parents(p.ID)

		// ------------------------------------------------
		// 规则一：
		// Gateway -> Condition -> Me
		// ------------------------------------------------
		if p.Type == "condition" &&
			hasType(grandParents, "parallel", "inclusion") {

			addEvent(&events, existEvent, "EventConcurrentRejectCleanup")
		}

		// ------------------------------------------------
		// 规则二：
		// Condition -> Gateway -> Me
		// ------------------------------------------------
		if isGateway(p.Type) &&
			hasType(grandParents, "condition") {

			addEvent(&events, existEvent, "EventGatewayConditionReject")
		}
	}

	return events
}

// createProxyWaitNode 创建代理等待节点（自动化节点）
// eventNodeId: 实际需要接收事件通知的网关节点 ID (parallel/inclusion)
func (l *logicFlow) createProxyWaitNode(prevNodeId, eventNodeId string) string {
	proxyNodeId := "proxy_" + prevNodeId + "_" + eventNodeId
	// 代理节点是一个通过型的自动化节点
	// 它接收 prevNodeId 的输入，然后自己可以瞬间完成
	// 完成时触发 eventNodeId 需要的事件
	n := model.Node{
		NodeID:           proxyNodeId,
		NodeName:         SysProxyNodeName,
		NodeType:         1, // User 节点
		PrevNodeIDs:      []string{prevNodeId},
		UserIDs:          []string{SysAutoUser},   // 标识为系统自动节点
		NodeStartEvents:  []string{"EventNotify"}, // User节点启动时触发Notify
		NodeEndEvents:    []string{},
		TaskFinishEvents: l.getPassEvents(proxyNodeId), // 这会根据 proxyNodeId 查找下级
	}

	// 修正：l.getPassEvents 依赖 l.Edges。我们的代理节点不在 Edge 表里。
	// 所以我们手动指定事件。
	// 我们知道 eventNodeId 是 parallel 或 inclusion。
	info := l.GetNodeInfo(eventNodeId)
	if info.Type == "parallel" {
		n.TaskFinishEvents = []string{"EventTaskParallelNodePass"}
	} else if info.Type == "inclusion" {
		n.TaskFinishEvents = []string{"EventTaskInclusionNodePass"}
	}

	l.NodeList = append(l.NodeList, n)
	return proxyNodeId
}

func (l *logicFlow) getProxyTargetNodeId(sourceId, targetId string) string {
	// 只有当 Source 是 Gateway (Cond/Parallel/Inc) 且 Target 也是 Gateway (Parallel/Inc) 时
	// 才使用了代理节点。
	// 注意：Condition节点后面也可以接代理，只要Target是Parallel/Inc。
	// Source节点的类型检查不需要在这里做，因为调用者本身就是 Condition/Parallel/Inclusion 方法。
	// 我们只需要检查 Target 类型。

	targetInfo := l.GetNodeInfo(targetId)
	if targetInfo.Type == "parallel" || targetInfo.Type == "inclusion" {
		return "proxy_" + sourceId + "_" + targetId
	}
	return targetId
}

func (l *logicFlow) findAndProxySrcNodes(node Node) []string {
	edgesSrc := l.FindSourceNodeId(node.ID)
	return slice.Map(edgesSrc, func(idx int, src Edge) string {
		// 检查上级节点类型，如果是 Condition / Parallel / Inclusion，则通过中间节点桥接
		srcNodeInfo := l.GetNodeInfo(src.SourceNodeId)
		if srcNodeInfo.Type == "condition" || srcNodeInfo.Type == "parallel" || srcNodeInfo.Type == "inclusion" {
			proxyNodeId := l.createProxyWaitNode(src.SourceNodeId, node.ID)
			return proxyNodeId
		}

		return src.SourceNodeId
	})
}

// FindPrevNodeIDs 查找上级节点的信息
func (l *logicFlow) FindPrevNodeIDs(id string) []string {
	edgesSrc := l.FindSourceNodeId(id)
	return slice.Map(edgesSrc, func(idx int, src Edge) string {
		return src.SourceNodeId
	})
}

// FindSourceNodeId 查找下级节点的连接线
func (l *logicFlow) FindSourceNodeId(id string) []Edge {
	return slice.FilterMap(l.Edges, func(idx int, src Edge) (Edge, bool) {
		if src.TargetNodeId == id {
			return src, true
		}

		return Edge{}, false
	})
}

// FindTargetNodeId 查找上级节点的连接线
func (l *logicFlow) FindTargetNodeId(id string) []Edge {
	return slice.FilterMap(l.Edges, func(idx int, src Edge) (Edge, bool) {
		if src.SourceNodeId == id {
			return src, true
		}

		return Edge{}, false
	})
}

func (l *logicFlow) ToTargetNode(nodeId string) string {
	for _, edge := range l.Edges {
		if edge.SourceNodeId == nodeId {
			return edge.TargetNodeId
		}
	}

	return ""
}

func (l *logicFlow) GetNodeInfo(nodeId string) Node {
	for _, node := range l.Nodes {
		if node.ID == nodeId {
			return node
		}
	}

	return Node{}
}

// ToEdgeProperty edge连线字段解析
func ToEdgeProperty(edges Edge) (EdgeProperty, error) {
	edgesJson, err := json.Marshal(edges.Properties)
	if err != nil {
		return EdgeProperty{}, err
	}

	var edgesProperty EdgeProperty
	err = json.Unmarshal(edgesJson, &edgesProperty)
	if err != nil {
		return EdgeProperty{}, err
	}

	return edgesProperty, nil
}

// ToNodeProperty node节点字段解析
func ToNodeProperty[T any](node Node) (T, error) {
	nodesJson, err := json.Marshal(node.Properties)
	if err != nil {
		return zeroValue[T](), err
	}

	var nodesProperty T
	err = json.Unmarshal(nodesJson, &nodesProperty)
	if err != nil {
		return zeroValue[T](), err
	}

	return nodesProperty, nil
}

func (l *logicFlow) toEdges() error {
	edgesJSON, err := json.Marshal(l.Workflow.FlowData.Edges)
	if err != nil {
		return err
	}

	var edges []Edge
	err = json.Unmarshal(edgesJSON, &edges)
	if err != nil {
		return err
	}

	l.Edges = edges
	return nil
}

func (l *logicFlow) toNodes() error {
	nodesJSON, err := json.Marshal(l.Workflow.FlowData.Nodes)
	if err != nil {
		return err
	}

	var nodes []Node
	err = json.Unmarshal(nodesJSON, &nodes)
	if err != nil {
		return err
	}

	l.Nodes = nodes
	return nil
}

func zeroValue[T any]() T {
	var zero T
	return zero
}
