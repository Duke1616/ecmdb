package easyflow

import (
	"encoding/json"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/ecodeclub/ekit/slice"
)

type logicFlow struct {
	Workflow Workflow
	Edges    []Edge
	Nodes    []Node

	// 后端存储结构体
	NodeList []model.Node
}

func NewLogicFlowToEngineConvert() ProcessEngineConvert {
	return &logicFlow{}
}

func (l *logicFlow) Deploy(workflow Workflow) (int, error) {
	l.Workflow = workflow
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
		}
	}

	// 发布流程
	process := model.Process{ProcessName: l.Workflow.Name, Source: "工单系统", RevokeEvents: []string{"EventRevoke"}, Nodes: l.NodeList}
	j, err := engine.JSONMarshal(process, false)
	if err != nil {
		return 0, err
	}

	return engine.ProcessSave(string(j), l.Workflow.Owner)
}

func (l *logicFlow) Start(node Node) {
	NodeName := "Start"
	property, _ := toNodeProperty[StartProperty](node)
	if property.Name != "" {
		NodeName = property.Name
	}
	n := model.Node{NodeID: node.ID, NodeName: NodeName,
		NodeType: 0, UserIDs: []string{"$starter"},
		NodeEndEvents: []string{"EventEnd"},
	}

	l.NodeList = append(l.NodeList, n)
}

func (l *logicFlow) End(node Node) {
	NodeName := "End"
	property, _ := toNodeProperty[EndProperty](node)
	if property.Name != "" {
		NodeName = property.Name
	}
	n := model.Node{NodeID: node.ID, NodeName: NodeName,
		NodeType: 3, PrevNodeIDs: l.FindPrevNodeIDs(node.ID),
		NodeStartEvents: []string{"EventNotify", "EventClose"},
	}
	l.NodeList = append(l.NodeList, n)
}

func (l *logicFlow) Condition(node Node) {
	// 获取所有判断条件的连接线
	edgesDst := l.FindTargetNodeId(node.ID)

	// 组合 conditions 条件
	conditions := slice.Map(edgesDst, func(idx int, src Edge) model.Condition {
		property, _ := toEdgeProperty(src)
		return model.Condition{
			Expression: property.Expression,
			NodeID:     src.TargetNodeId,
		}
	})

	// 拼接网关
	GwCondition := model.HybridGateway{Conditions: conditions, InevitableNodes: []string{}, WaitForAllPrevNode: 0}

	// node 节点录入
	property, _ := toNodeProperty[ConditionProperty](node)
	n := model.Node{NodeID: node.ID, NodeName: property.Name,
		NodeType: 2, GWConfig: GwCondition,
		PrevNodeIDs: l.FindPrevNodeIDs(node.ID),
	}

	l.NodeList = append(l.NodeList, n)
}

func (l *logicFlow) User(node Node) {
	// node 节点录入
	property, _ := toNodeProperty[UserProperty](node)
	n := model.Node{NodeID: node.ID, NodeName: property.Name,
		NodeType: 1, UserIDs: []string{property.Approved},
		PrevNodeIDs: l.FindPrevNodeIDs(node.ID),
	}

	l.NodeList = append(l.NodeList, n)
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

// edge连线字段解析
func toEdgeProperty(edges Edge) (EdgeProperty, error) {
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

// node节点字段解析
func toNodeProperty[T any](node Node) (T, error) {
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
