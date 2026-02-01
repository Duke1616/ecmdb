package service

import (
	"context"

	"github.com/Bunny3th/easy-workflow/workflow/engine"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/engine/internal/domain"
	"github.com/Duke1616/ecmdb/internal/engine/internal/repository"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	// ListTodoTasks 查看todo任务
	ListTodoTasks(ctx context.Context, userId, processName string, sortByAse bool, offset, limit int) (
		[]domain.Instance, int64, error)

	ListByStartUser(ctx context.Context, userId, processName string, offset, limit int) (
		[]domain.Instance, int64, error)
	// TaskRecord 工单任务变更记录
	TaskRecord(ctx context.Context, processInstId, offset, limit int) ([]model.Task, int64, error)
	IsReject(ctx context.Context, taskId int) (bool, error)
	// GetTasksByCurrentNodeId 获取当前节点下的所有任务
	GetTasksByCurrentNodeId(ctx context.Context, processInstId int, currentNodeId string) ([]model.Task, error)
	// UpdateIsFinishedByPreNodeId 系统修改 finished 状态
	UpdateIsFinishedByPreNodeId(ctx context.Context, nodeId string, status int, comment string) error
	// ForceUpdateIsFinishedByPreNodeId 强制清理指定节点下的所有任务（包括已完成的）
	ForceUpdateIsFinishedByPreNodeId(ctx context.Context, nodeId string, status int, comment string) error
	// ForceUpdateIsFinishedByNodeId 强制清理指定节点ID的所有任务（包括已完成的）
	ForceUpdateIsFinishedByNodeId(ctx context.Context, nodeId string, status int, comment string) error
	// Pass 通过
	Pass(ctx context.Context, taskId int, comment string) error
	// ListPendingStepsOfMyTask 列出我的任务待处理步骤
	ListPendingStepsOfMyTask(ctx context.Context, processInstIds []int, starter string) ([]domain.Instance, error)
	// GetAutomationTask 获取自动化完成任务
	GetAutomationTask(ctx context.Context, currentNodeId string, processInstId int) (model.Task, error)
	// GetTasksByInstUsers 获取指定流程 + 用户的任务
	GetTasksByInstUsers(ctx context.Context, processInstId int, userIds []string) ([]model.Task, error)
	// GetOrderIdByVariable 获取工单ID，进行流程绑定
	GetOrderIdByVariable(ctx context.Context, processInstId int) (string, error)
	// Upstream 获取所有上游节点
	Upstream(ctx context.Context, taskId int) ([]model.Node, error)
	// TaskInfo 获取任务详情
	TaskInfo(ctx context.Context, taskId int) (model.Task, error)
	// GetProxyPrevNodeID 获取代理转发的节点ID
	GetProxyPrevNodeID(ctx context.Context, prevNodeID string) (string, error)
	// GetProxyNodeID 获取代理转发的节点ID
	GetProxyNodeID(ctx context.Context, prevNodeID string) (string, error)
	// GetProxyNodeByProcessInstId 通过流程实例ID获取 proxy 节点ID
	GetProxyNodeByProcessInstId(ctx context.Context, processInstId int) (string, error)
	// GetProxyTaskByProcessInstId 通过流程实例ID获取 proxy 节点完整信息
	GetProxyTaskByProcessInstId(ctx context.Context, processInstId int) (model.Task, error)
	// DeleteProxyNodeByNodeId 删除指定 proxy 节点任务记录
	DeleteProxyNodeByNodeId(ctx context.Context, nodeId string) error
	// UpdateTaskPrevNodeID 修改任务节点ID
	UpdateTaskPrevNodeID(ctx context.Context, taskId int, prevNodeId string) error
	// GetTraversedEdges 获取已流转的边
	GetTraversedEdges(ctx context.Context, processInstId, processId int, status uint8) (map[string][]string, error)
}

type service struct {
	repo repository.ProcessEngineRepository
}

func (s *service) GetProxyPrevNodeID(ctx context.Context, prevNodeID string) (string, error) {
	procTask, err := s.repo.GetProxyNodeID(ctx, prevNodeID)
	return procTask.PrevNodeID, err
}

func (s *service) GetProxyNodeID(ctx context.Context, prevNodeID string) (string, error) {
	procTask, err := s.repo.GetProxyNodeID(ctx, prevNodeID)
	return procTask.NodeID, err
}

func (s *service) GetProxyNodeByProcessInstId(ctx context.Context, processInstId int) (string, error) {
	procTask, err := s.repo.GetProxyNodeByProcessInstId(ctx, processInstId)
	// NOTE: 返回 proxy 节点的 NodeID，用于后续更新状态
	return procTask.NodeID, err
}

func (s *service) GetProxyTaskByProcessInstId(ctx context.Context, processInstId int) (model.Task, error) {
	// NOTE: 返回完整的 Task 对象，包含 PrevNodeID 等信息
	return s.repo.GetProxyNodeByProcessInstId(ctx, processInstId)
}

func (s *service) DeleteProxyNodeByNodeId(ctx context.Context, nodeId string) error {
	return s.repo.DeleteProxyNodeByNodeId(ctx, nodeId)
}

func (s *service) TaskInfo(ctx context.Context, taskId int) (model.Task, error) {
	return engine.GetTaskInfo(taskId)
}

func (s *service) GetTasksByCurrentNodeId(ctx context.Context, processInstId int, currentNodeId string) ([]model.Task, error) {
	return s.repo.GetTasksByCurrentNodeId(ctx, processInstId, currentNodeId)
}

func (s *service) Upstream(ctx context.Context, taskId int) ([]model.Node, error) {
	return engine.TaskUpstreamNodeList(taskId)
}

func (s *service) GetOrderIdByVariable(ctx context.Context, processInstId int) (string, error) {
	return s.repo.GetOrderIdByVariable(ctx, processInstId)
}

func (s *service) GetTasksByInstUsers(ctx context.Context, processInstId int, userIds []string) ([]model.Task, error) {
	return s.repo.GetTasksByInstUsers(ctx, processInstId, userIds)
}

func (s *service) GetAutomationTask(ctx context.Context, currentNodeId string, processInstId int) (model.Task, error) {
	return s.repo.GetAutomationTask(ctx, currentNodeId, processInstId)
}

func (s *service) ListPendingStepsOfMyTask(ctx context.Context, processInstIds []int, starter string) (
	[]domain.Instance, error) {
	return s.repo.ListTasksByProcInstIds(ctx, processInstIds, starter)
}

func (s *service) IsReject(ctx context.Context, taskId int) (bool, error) {
	total, err := s.repo.CountReject(ctx, taskId)

	if total >= 1 {
		return true, err
	}

	return false, err
}

func (s *service) UpdateIsFinishedByPreNodeId(ctx context.Context, nodeId string, status int, comment string) error {
	return s.repo.UpdateIsFinishedByPreNodeId(ctx, nodeId, status, comment)
}

// ForceUpdateIsFinishedByPreNodeId 强制清理指定节点下的所有任务（包括已完成的）
func (s *service) ForceUpdateIsFinishedByPreNodeId(ctx context.Context, nodeId string, status int, comment string) error {
	return s.repo.ForceUpdateIsFinishedByPreNodeId(ctx, nodeId, status, comment)
}

// ForceUpdateIsFinishedByNodeId 强制清理指定节点ID的所有任务（包括已完成的）
func (s *service) ForceUpdateIsFinishedByNodeId(ctx context.Context, nodeId string, status int, comment string) error {
	return s.repo.ForceUpdateIsFinishedByNodeId(ctx, nodeId, status, comment)
}

func (s *service) Pass(ctx context.Context, taskId int, comment string) error {
	return engine.TaskPass(taskId, comment, "", false)
}

func (s *service) UpdateTaskPrevNodeID(ctx context.Context, taskId int, prevNodeId string) error {
	return s.repo.UpdateTaskPrevNodeID(ctx, taskId, prevNodeId)
}

func (s *service) TaskRecord(ctx context.Context, processInstId, offset, limit int) ([]model.Task, int64, error) {
	var (
		eg      errgroup.Group
		records []model.Task
		total   int64
	)
	eg.Go(func() error {
		var err error
		records, err = s.repo.ListTaskRecord(ctx, processInstId, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountTaskRecord(ctx, processInstId)
		return err
	})
	if err := eg.Wait(); err != nil {
		return records, total, err
	}
	return records, total, nil
}

func NewService(repo repository.ProcessEngineRepository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) ListTodoTasks(ctx context.Context, userId, processName string, sortByAse bool, offset, limit int) (
	[]domain.Instance, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Instance
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.TodoList(userId, processName, sortByAse, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountTodo(ctx, userId, processName)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) ListByStartUser(ctx context.Context, userId, processName string, offset,
	limit int) ([]domain.Instance, int64, error) {

	var (
		eg    errgroup.Group
		ts    []domain.Instance
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListStartUser(ctx, userId, processName, offset, limit)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountStartUser(ctx, userId, processName)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) GetTraversedEdges(ctx context.Context, processInstId, processId int, status uint8) (map[string][]string, error) {
	record, _, err := s.TaskRecord(ctx, processInstId, 0, 1000)
	if err != nil {
		return nil, err
	}

	define, err := engine.GetProcessDefine(processId)
	if err != nil {
		return nil, err
	}
	// 找到 Root ID
	var rootID string
	nodesMap := slice.ToMap(define.Nodes, func(element model.Node) string {
		if element.NodeType == model.RootNode {
			rootID = element.NodeID
		}
		return element.NodeID
	})
	if rootID == "" && len(record) > 0 {
		rootID = record[0].NodeID // Fallback
	}

	// 1. 构建有效图 (Adjacency Map)，处理 Proxy 穿透
	// Key: SourceID, Value: List of TargetIDs
	effectiveGraph := s.buildEffectiveGraph(nodesMap)

	// 2. 状态重放
	// litEdges: 当前点亮的边 map[source] -> []target
	litEdges := make(map[string][]string)

	for _, task := range record {
		// 2.1 递归重置 (Recursive Reset)
		// 如果流程重新进入了该节点，清除它及后续网关的激活状态
		s.recursiveReset(task.NodeID, litEdges, nodesMap)

		// 2.2 点亮 Incoming 边: Prev -> Current
		// 驳回节点 (Status=2) 的 Incoming 是回退线，通常不画
		if task.Status != 2 && task.PrevNodeID != "" {
			// 将 DB 里的 Proxy ID "翻译" 回逻辑上的前置 ID
			logicalPrevID := s.getLogicalPrevID(task.PrevNodeID, nodesMap)

			// 搜索从 Prev 到 Current 的路径 (允许穿透中间的网关)
			//这解决了 DB 记录跳过网关 (Prev=提交人, Current=李四, 中间有网关) 的问题
			path := s.findPathThroughGateways(logicalPrevID, task.NodeID, effectiveGraph, nodesMap)
			if len(path) > 1 {
				for i := 0; i < len(path)-1; i++ {
					uniqueAppend(litEdges, path[i], path[i+1])
				}
			}
		}

		// 2.3 前向 Look-ahead (处理已完成任务指向的后续节点，如网关)
		if task.IsFinished == 1 && task.Status != 2 {
			nextIDs := effectiveGraph[task.NodeID]
			for _, nextID := range nextIDs {
				// 记录边，并递归点亮后续的纯网关路径
				s.lightUpForward(task.NodeID, nextID, litEdges, nodesMap, effectiveGraph)
			}
		}
	}

	// 3. 可达性过滤 (Reachability Filter)
	// 从 Root 开始 BFS，清理因重置而断开的孤岛路径
	finalEdges := make(map[string][]string)

	if rootID != "" {
		queue := []string{rootID}
		visitedNode := make(map[string]bool)
		visitedNode[rootID] = true

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			targets := litEdges[curr]
			if len(targets) > 0 {
				finalEdges[curr] = targets
				for _, t := range targets {
					if !visitedNode[t] {
						visitedNode[t] = true
						queue = append(queue, t)
					}
				}
			}
		}
	} else {
		finalEdges = litEdges
	}

	return finalEdges, nil
}

// findPathThroughGateways 寻找从 start 到 end 的路径，且中间节点必须是网关
func (s *service) findPathThroughGateways(startID, endID string,
	effectiveGraph map[string][]string, nodesMap map[string]model.Node) []string {

	type path struct {
		nodes []string
	}

	queue := []path{{nodes: []string{startID}}}
	// 注意：visited 可以防止在一次搜索中产生环路，但如果是 DAG 其实没问题。
	// 为了简单，我们记录 visited。
	visited := map[string]bool{startID: true}

	for len(queue) > 0 {
		currPath := queue[0]
		queue = queue[1:]

		currNodeID := currPath.nodes[len(currPath.nodes)-1]

		if currNodeID == endID {
			return currPath.nodes
		}

		// 限制深度
		if len(currPath.nodes) > 20 {
			continue
		}

		nextIDs := effectiveGraph[currNodeID]
		for _, nextID := range nextIDs {
			isTarget := nextID == endID
			isGateway := false
			if node, ok := nodesMap[nextID]; ok && node.NodeType == model.GateWayNode {
				isGateway = true
			}

			if (isTarget || isGateway) && !visited[nextID] {
				visited[nextID] = true
				newNodes := make([]string, len(currPath.nodes))
				copy(newNodes, currPath.nodes)
				newNodes = append(newNodes, nextID)

				if isTarget {
					return newNodes
				}
				queue = append(queue, path{nodes: newNodes})
			}
		}
	}
	return nil
}

// recursiveReset 清除节点及其后续(如果是网关)的激活边
func (s *service) recursiveReset(nodeID string, litEdges map[string][]string, nodesMap map[string]model.Node) {
	targets, ok := litEdges[nodeID]
	if !ok {
		return
	}
	delete(litEdges, nodeID)

	for _, tid := range targets {
		tNode, exists := nodesMap[tid]
		if exists && tNode.NodeType == model.GateWayNode {
			s.recursiveReset(tid, litEdges, nodesMap)
		}
	}
}

// lightUpForward 递归点亮前向路径 (穿透网关)
func (s *service) lightUpForward(sourceID, targetID string, litEdges map[string][]string,
	nodesMap map[string]model.Node, effectiveGraph map[string][]string) {

	uniqueAppend(litEdges, sourceID, targetID)

	// 如果目标是网关，继续递归点亮它的下一级
	targetNode, ok := nodesMap[targetID]
	if ok && targetNode.NodeType == model.GateWayNode {
		nextIDs := effectiveGraph[targetID]

		inDegree := len(targetNode.PrevNodeIDs)
		outDegree := len(nextIDs)
		waitType := targetNode.GWConfig.WaitForAllPrevNode

		// 规则 1. 并行网关 (Wait=1):
		//    - 汇聚 (In>1): 必须等待所有前置到达 -> 停止前探
		//    - 分支 (In=1): 并行同时触发 -> 允许穿透
		if waitType == 1 && inDegree > 1 {
			return
		}

		// 规则 2. 包容网关 (Wait=0):
		//    - 汇聚: 需要等待所有Active分支 -> 停止前探
		//    - 分支: 往往带有条件，不一定全走 -> 停止前探
		if waitType == 0 {
			return
		}

		// 规则 3. 排他/条件网关 (Wait=3):
		//    - 汇聚: 不等待 (XOR Merge 只要有一个到达即走) -> 允许穿透
		//    - 分支 (Out>1): 互斥路径，只走一条 -> 停止前探 (由后续生成的 Record 回溯点亮)
		if waitType == 3 && outDegree > 1 {
			return
		}

		for _, nextID := range nextIDs {
			s.lightUpForward(targetID, nextID, litEdges, nodesMap, effectiveGraph)
		}
	}
}

// getLogicalPrevID 递归查找逻辑前置（跳过 Proxy）
func (s *service) getLogicalPrevID(rawPrevID string, nodesMap map[string]model.Node) string {
	node, ok := nodesMap[rawPrevID]
	if !ok {
		return rawPrevID
	}

	if s.isProxyNode(node) {
		if len(node.PrevNodeIDs) > 0 {
			// 递归查找，直到找到非 Proxy 节点
			return s.getLogicalPrevID(node.PrevNodeIDs[0], nodesMap)
		}
	}
	return rawPrevID
}

// buildEffectiveGraph 构建忽略 Proxy 的真实节点邻接表
func (s *service) buildEffectiveGraph(nodesMap map[string]model.Node) map[string][]string {
	graph := make(map[string][]string)

	for id, node := range nodesMap {
		if s.isProxyNode(node) {
			continue
		}

		realPrevs := s.resolveRealPrevs(node, nodesMap)
		for _, prevID := range realPrevs {
			// Graph: Prev -> ID
			uniqueAppend(graph, prevID, id)
		}
	}
	return graph
}

// 辅助：Resolve real prevs (skipping proxy)
func (s *service) resolveRealPrevs(node model.Node, nodesMap map[string]model.Node) []string {
	var result []string
	for _, prevID := range node.PrevNodeIDs {
		prevNode, exists := nodesMap[prevID]
		if !exists {
			continue
		}

		if s.isProxyNode(prevNode) {
			result = append(result, s.resolveRealPrevs(prevNode, nodesMap)...)
		} else {
			result = append(result, prevID)
		}
	}
	return result
}

func uniqueAppend(m map[string][]string, key, val string) {
	if !slice.Contains(m[key], val) {
		m[key] = append(m[key], val)
	}
}

func (s *service) isProxyNode(node model.Node) bool {
	for _, uid := range node.UserIDs {
		if uid == easyflow.SysAutoUser {
			return true
		}
	}
	return false
}
