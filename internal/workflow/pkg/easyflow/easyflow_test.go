package easyflow

import (
	"testing"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogicFlow_Deploy(t *testing.T) {
	testCases := []struct {
		name     string
		workflow Workflow
		verify   func(t *testing.T, nodes []model.Node)
	}{
		{
			name: "基础流程转换: Start -> User -> End",
			workflow: Workflow{
				Name:  "基础流程",
				Owner: "admin",
				FlowData: LogicFlow{
					Nodes: []map[string]interface{}{
						{"id": "node_start", "type": "start", "properties": map[string]interface{}{"name": "开始"}},
						{"id": "node_user", "type": "user", "properties": map[string]interface{}{"name": "审批", "approved": []string{"manager"}}},
						{"id": "node_end", "type": "end", "properties": map[string]interface{}{"name": "结束"}},
					},
					Edges: []map[string]interface{}{
						{"id": "edge1", "sourceNodeId": "node_start", "targetNodeId": "node_user"},
						{"id": "edge2", "sourceNodeId": "node_user", "targetNodeId": "node_end"},
					},
				},
			},
			verify: func(t *testing.T, nodes []model.Node) {
				require.Len(t, nodes, 3)

				startNode, _ := findNode(nodes, "node_start")
				require.NotNil(t, startNode)
				assert.Equal(t, model.NodeType(0), startNode.NodeType)

				userNode, _ := findNode(nodes, "node_user")
				require.NotNil(t, userNode)
				assert.Contains(t, userNode.UserIDs, "manager")

				endNode, _ := findNode(nodes, "node_end")
				require.NotNil(t, endNode)
				assert.Equal(t, model.NodeType(3), endNode.NodeType)
			},
		},
		{
			name: "复杂过程转换: 网关与代理节点",
			workflow: Workflow{
				Name:  "网关流程",
				Owner: "admin",
				FlowData: LogicFlow{
					Nodes: []map[string]interface{}{
						{"id": "n1", "type": "start"},
						{"id": "n2", "type": "condition", "properties": map[string]interface{}{"name": "条件"}},
						{"id": "n3", "type": "parallel", "properties": map[string]interface{}{"name": "并行汇聚"}},
						{"id": "n4", "type": "end"},
					},
					Edges: []map[string]interface{}{
						{"id": "e1", "sourceNodeId": "n1", "targetNodeId": "n2"},
						{"id": "e2", "sourceNodeId": "n2", "targetNodeId": "n3", "properties": map[string]interface{}{"expression": "a == 1"}},
						{"id": "e3", "sourceNodeId": "n3", "targetNodeId": "n4"},
					},
				},
			},
			verify: func(t *testing.T, nodes []model.Node) {
				// n1, n2, n3, n4 + 1个proxy节点
				assert.Len(t, nodes, 5)

				proxyID := "proxy_n2_n3"
				proxyNode, ok := findNode(nodes, proxyID)
				require.True(t, ok)
				assert.Equal(t, SysProxyNodeName, proxyNode.NodeName)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			converter := NewDefaultConverter()
			converter.Register(&StartNodeHandler{})
			converter.Register(&EndNodeHandler{})
			converter.Register(&UserNodeHandler{})
			converter.Register(&ParallelHandler{})
			converter.Register(&SelectiveHandler{})
			converter.Register(&ConditionHandler{})

			process, err := converter.Convert(tc.workflow)
			require.NoError(t, err)

			tc.verify(t, process.Nodes)
		})
	}
}

func findNode(nodes []model.Node, id string) (model.Node, bool) {
	for _, n := range nodes {
		if n.NodeID == id {
			return n, true
		}
	}
	return model.Node{}, false
}

func TestUserProperty_NormalizeAssignees(t *testing.T) {
	testCases := []struct {
		name     string
		property UserProperty
		want     []Assignee
	}{
		{
			name: "新版数据模式",
			property: UserProperty{
				Assignees: []Assignee{
					{Rule: APPOINT, Values: []string{"user1", "user2"}},
				},
			},
			want: []Assignee{
				{Rule: APPOINT, Values: []string{"user1", "user2"}},
			},
		},
		{
			name: "老版本模式-模板字段",
			property: UserProperty{
				Rule:          TEMPLATE,
				TemplateField: "manager",
			},
			want: []Assignee{
				{Rule: TEMPLATE, Values: []string{"manager"}},
			},
		},
		{
			name: "老版本模式-指定人",
			property: UserProperty{
				Rule:     APPOINT,
				Approved: []string{"user3"},
			},
			want: []Assignee{
				{Rule: APPOINT, Values: []string{"user3"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.property.NormalizeAssignees()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestUpdateEdgeProperties(t *testing.T) {
	edgeMap := map[string][]string{
		"node1": {"node2", "node3"},
	}

	t.Run("SkipCase", func(t *testing.T) {
		edges := []Edge{
			{SourceNodeId: "node1", TargetNodeId: "node2", Properties: map[string]interface{}{}},
			{SourceNodeId: "node1", TargetNodeId: "node3", Properties: map[string]interface{}{}},
		}
		nodeStatusMap := map[string]int{
			"node1": 5, "node2": 2, "node3": 5,
		}
		updatedEdges := UpdateEdgeProperties(edges, edgeMap, nodeStatusMap)
		for _, e := range updatedEdges {
			props := e.Properties.(map[string]interface{})
			assert.True(t, props["is_skipped"].(bool))
		}
	})

	t.Run("PassCase", func(t *testing.T) {
		nodeStatusMapPass := map[string]int{"node1": 1, "node2": 1}
		edgesPass := []Edge{{SourceNodeId: "node1", TargetNodeId: "node2", Properties: map[string]interface{}{}}}
		updated := UpdateEdgeProperties(edgesPass, edgeMap, nodeStatusMapPass)
		props := updated[0].Properties.(map[string]interface{})
		assert.True(t, props["is_pass"].(bool))
	})
}
