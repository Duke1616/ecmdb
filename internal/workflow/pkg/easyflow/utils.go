package easyflow

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mitchellh/mapstructure"
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

	// 增加排序逻辑：解决重叠线条下，“通过”或“跳过”状态的线条被默认（灰色）线条覆盖的问题。
	// 通过提高 Z-Index 逻辑（在 LogicFlow 中表现为数组顺序越后，层级越高），将点亮的线条放到数组末尾。
	sort.Slice(edges, func(i, j int) bool {
		pi, _ := edges[i].Properties.(map[string]interface{})
		pj, _ := edges[j].Properties.(map[string]interface{})

		// 定义渲染优先级：普通(0) < 已跳过(1) < 已通过(2)
		getPriority := func(p map[string]interface{}) int {
			if p == nil {
				return 0
			}
			if p["is_pass"] == true {
				return 2
			}
			if p["is_skipped"] == true {
				return 1
			}
			return 0
		}

		return getPriority(pi) < getPriority(pj)
	})

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

// ToEdgeProperty edge连线字段解析
func ToEdgeProperty(edges Edge) (EdgeProperty, error) {
	var property EdgeProperty
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &property,
		TagName: "json",
	})
	if err != nil {
		return EdgeProperty{}, err
	}

	if err = decoder.Decode(edges.Properties); err != nil {
		return EdgeProperty{}, err
	}

	return property, nil
}

// ToNodeProperty node节点字段解析
func ToNodeProperty[T any](node Node) (T, error) {
	var property T
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &property,
		TagName: "json",
	})
	if err != nil {
		return zeroValue[T](), err
	}

	if err = decoder.Decode(node.Properties); err != nil {
		return zeroValue[T](), err
	}

	return property, nil
}

func zeroValue[T any]() T {
	var zero T
	return zero
}
