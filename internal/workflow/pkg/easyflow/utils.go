package easyflow

import (
	"encoding/json"
	"fmt"
)

// UpdateEdgeProperties 更新边的属性，用于设置 Pass/Skip 状态
func UpdateEdgeProperties(edges []Edge, edgeMap map[string][]string, nodeStatusMap map[string]int) []Edge {
	for i, edge := range edges {
		properties, ok := edge.Properties.(map[string]interface{})
		if !ok {
			properties = make(map[string]interface{})
		}

		targetNodeIDs, ok := edgeMap[edge.SourceNodeId]
		if !ok {
			continue
		}

		// 检查当前边的 Target 是否在 targets 列表中
		isMatched := false
		for _, tid := range targetNodeIDs {
			if tid == edge.TargetNodeId {
				isMatched = true
				break
			}
		}

		if !isMatched {
			continue
		}

		// 检查节点状态 (Source 或 Target 为 status=5)
		sourceStatus := nodeStatusMap[edge.SourceNodeId]
		targetStatus := nodeStatusMap[edge.TargetNodeId]

		if sourceStatus == 5 || targetStatus == 5 {
			// 被跳过的分支，标记为 is_skipped=true
			properties["is_skipped"] = true
		} else {
			// 正常通过的分支，标记为 is_pass=true
			properties["is_pass"] = true
		}

		edges[i].Properties = properties
	}

	return edges
}

// GetJsCode 供飞书或前端使用的通用 JSON 代码生成函数
// 仅根据 Workflow 对象和状态映射生成 window.__DATA__ 代码
func GetJsCode(wfFlowData LogicFlow, edgeMap map[string][]string, nodeStatusMap map[string]int) (string, error) {
	edgesJSON, err := json.Marshal(wfFlowData.Edges)
	if err != nil {
		return "", err
	}

	var edges []Edge
	err = json.Unmarshal(edgesJSON, &edges)
	if err != nil {
		return "", err
	}

	// 更新属性
	edges = UpdateEdgeProperties(edges, edgeMap, nodeStatusMap)

	// 重建 EasyFlow 数据
	var edgesMap []map[string]interface{}
	for _, edge := range edges {
		edgesMap = append(edgesMap, EdgeToMap(edge))
	}

	wfFlowData.Edges = edgesMap
	easyFlowData, err := json.Marshal(wfFlowData)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`window.__DATA__ = %s;`, easyFlowData), nil
}

func EdgeToMap(edge Edge) map[string]interface{} {
	return map[string]interface{}{
		"type":         edge.Type,
		"sourceNodeId": edge.SourceNodeId,
		"targetNodeId": edge.TargetNodeId,
		"properties":   edge.Properties,
		"id":           edge.ID,
		"startPoint":   edge.StartPoint,
		"endPoint":     edge.EndPoint,
		"pointsList":   edge.PointsList,
		"text":         edge.Text,
	}
}
