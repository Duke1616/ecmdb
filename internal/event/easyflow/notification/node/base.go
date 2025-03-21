package node

import (
	"encoding/json"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/workflow"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
)

// 比对自动化任务节点消息发送模式，如果存在则为 True， 不存在则返回 False
func containsAutoNotifyMethod(notifyMethod []int64, target int64) bool {
	for _, item := range notifyMethod {
		if item == target {
			return true
		}
	}
	return false
}

// 返回节点，但是不需要返回node的情况
func getNodeProperty[T any](wf workflow.Workflow, currentNodeId string) (T, error) {
	nodes, err := unmarshal(wf)
	if err != nil {
		return *new(T), err
	}
	return getProperty[T](nodes, currentNodeId)
}

// 解析数据
func unmarshal(wf workflow.Workflow) ([]easyflow.Node, error) {
	nodesJSON, err := json.Marshal(wf.FlowData.Nodes)
	if err != nil {
		return nil, err
	}
	var nodes []easyflow.Node
	err = json.Unmarshal(nodesJSON, &nodes)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// 统一解析不同的节点属性
func getProperty[T any](nodes []easyflow.Node, currentNodeId string) (T, error) {
	var property T
	for _, node := range nodes {
		if node.ID == currentNodeId {
			return easyflow.ToNodeProperty[T](node)
		}
	}
	return property, fmt.Errorf("未找到节点 %s 的属性", currentNodeId)
}
