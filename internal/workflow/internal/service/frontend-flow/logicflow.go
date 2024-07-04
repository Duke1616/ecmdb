package frontend_flow

import (
	"encoding/json"
	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/domain"
	"github.com/ecodeclub/ekit/slice"
)

type loginFlow struct {
	flow  domain.Workflow
	Edges []domain.Edge
	Nodes []domain.Node

	// 后端存储结构体
	NodeList []model.Node
}

func NewFrontendFlow(req domain.Workflow) FrontendFlow {
	return &loginFlow{
		flow: req,
	}
}

func (l *loginFlow) Deploy() (int, error) {
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
	process := model.Process{ProcessName: l.flow.Name, Source: "工单系统", RevokeEvents: []string{"EVENT_REVOKE"}, Nodes: l.NodeList}
	j, err := engine.JSONMarshal(process, false)
	if err != nil {
		return 0, err
	}

	return engine.ProcessSave(string(j), l.flow.Owner)
}

func (l *loginFlow) Start(node domain.Node) {
	n := model.Node{NodeID: node.ID, NodeName: "Start",
		NodeType: 0, UserIDs: []string{"$starter"},
		NodeEndEvents: []string{"EVENT_END"},
	}

	l.NodeList = append(l.NodeList, n)
}

func (l *loginFlow) End(node domain.Node) {
	n := model.Node{NodeID: node.ID, NodeName: "End",
		NodeType: 3, PrevNodeIDs: []string{"GW-Day", "Boss"},
		NodeStartEvents: []string{"MyEvent_Notify"}}

	l.NodeList = append(l.NodeList, n)
}

func (l *loginFlow) Condition(node domain.Node) {
	// 获取所有判断条件的连接线
	edgesDst := l.FindTargetNodeId(node.ID)

	// 组合 conditions 条件
	conditions := slice.Map(edgesDst, func(idx int, src domain.Edge) model.Condition {
		property, _ := toEdgeProperty(src)
		return model.Condition{
			Expression: property.Expression,
			NodeID:     src.TargetNodeId,
		}
	})

	// 拼接网关
	GwCondition := model.HybridGateway{Conditions: conditions, InevitableNodes: []string{}, WaitForAllPrevNode: 0}

	// node 节点录入
	property, _ := toNodeProperty[domain.ConditionProperty](node)
	n := model.Node{NodeID: node.ID, NodeName: property.Name,
		NodeType: 2, GWConfig: GwCondition,
		PrevNodeIDs: l.FindPrevNodeIDs(node.ID),
	}

	l.NodeList = append(l.NodeList, n)
}

func (l *loginFlow) User(node domain.Node) {
	// node 节点录入
	property, _ := toNodeProperty[domain.UserProperty](node)
	n := model.Node{NodeID: node.ID, NodeName: property.Name,
		NodeType: 1, UserIDs: []string{property.Approved},
		PrevNodeIDs: l.FindPrevNodeIDs(node.ID),
	}

	l.NodeList = append(l.NodeList, n)
}

// FindPrevNodeIDs 查找上级节点的信息
func (l *loginFlow) FindPrevNodeIDs(id string) []string {
	edgesSrc := l.FindSourceNodeId(id)
	return slice.Map(edgesSrc, func(idx int, src domain.Edge) string {
		return src.SourceNodeId
	})
}

// FindSourceNodeId 查找下级节点的连接线
func (l *loginFlow) FindSourceNodeId(id string) []domain.Edge {
	return slice.FilterMap(l.Edges, func(idx int, src domain.Edge) (domain.Edge, bool) {
		if src.TargetNodeId == id {
			return src, true
		}

		return domain.Edge{}, false
	})
}

// FindTargetNodeId 查找上级节点的连接线
func (l *loginFlow) FindTargetNodeId(id string) []domain.Edge {
	return slice.FilterMap(l.Edges, func(idx int, src domain.Edge) (domain.Edge, bool) {
		if src.SourceNodeId == id {
			return src, true
		}

		return domain.Edge{}, false
	})
}

// edge连线字段解析
func toEdgeProperty(edges domain.Edge) (domain.EdgeProperty, error) {
	edgesJson, err := json.Marshal(edges.Properties)
	if err != nil {
		return domain.EdgeProperty{}, err
	}

	var edgesProperty domain.EdgeProperty
	err = json.Unmarshal(edgesJson, &edgesProperty)
	if err != nil {
		return domain.EdgeProperty{}, err
	}

	return edgesProperty, nil
}

// node节点字段解析
func toNodeProperty[T any](node domain.Node) (T, error) {
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

func (l *loginFlow) toEdges() error {
	edgesJSON, err := json.Marshal(l.flow.FlowData.Edges)
	if err != nil {
		return err
	}

	var edges []domain.Edge
	err = json.Unmarshal(edgesJSON, &edges)
	if err != nil {
		return err
	}

	l.Edges = edges
	return nil
}

func (l *loginFlow) toNodes() error {
	nodesJSON, err := json.Marshal(l.flow.FlowData.Nodes)
	if err != nil {
		return err
	}

	var nodes []domain.Node
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
