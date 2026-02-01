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
	var endNodeId string
	nodesMap := slice.ToMap(define.Nodes, func(element model.Node) string {
		if element.NodeType == model.EndNode {
			endNodeId = element.NodeID
		}
		return element.NodeID
	})

	filterRecord := slice.FilterMap(record, func(idx int, src model.Task) (model.Task, bool) {
		if src.Status == 2 {
			return model.Task{}, false
		}
		return src, true
	})
	recordMap := slice.ToMap(filterRecord, func(element model.Task) string {
		return element.NodeID
	})

	edges := make(map[string][]string)
	visited := make(map[string]bool)

	// 构建 NextNodesMap 用于前向查找
	nextNodesMap := make(map[string][]string)
	for id, node := range nodesMap {
		for _, prev := range node.PrevNodeIDs {
			nextNodesMap[prev] = append(nextNodesMap[prev], id)
		}
	}

	for _, task := range filterRecord {
		// 1. 回溯法 (原有逻辑)
		s.processNode(task.NodeID, nodesMap, recordMap, edges, visited)

		// 2. 前向法 (新增逻辑): 如果任务已完成，尝试点亮它指向下一节点的线
		// 这主要解决：后续节点是网关（Gateway）且因为等待其他分支而尚未生成 Task Record 的情况
		if task.IsFinished == 1 {
			nextIDs := nextNodesMap[task.NodeID]
			for _, nextID := range nextIDs {
				s.processForwardEdge(task.NodeID, nextID, nodesMap, nextNodesMap, edges)
			}
		}
	}

	if status == 3 {
		s.processNode(endNodeId, nodesMap, recordMap, edges, visited)
	}

	return edges, nil
}

func (s *service) processForwardEdge(sourceID, targetID string, nodesMap map[string]model.Node,
	nextNodesMap map[string][]string, edges map[string][]string) {

	targetNode, exists := nodesMap[targetID]
	if !exists {
		return
	}

	// Proxy 穿透处理
	if s.isProxyNode(targetNode) {
		// 如果目标是 Proxy，则跳过它，继续找 Proxy 的下一个节点
		proxyNextIDs := nextNodesMap[targetID]
		for _, proxyNextID := range proxyNextIDs {
			// 递归穿透 (sourceID 不变，target 变为 proxy 的 next)
			s.processForwardEdge(sourceID, proxyNextID, nodesMap, nextNodesMap, edges)
		}
		return
	}

	// 记录边：Source -> Target
	// 避免重复添加
	if !slice.Contains(edges[sourceID], targetID) {
		edges[sourceID] = append(edges[sourceID], targetID)
	}
}

func (s *service) processNode(nodeID string, nodesMap map[string]model.Node,
	recordMap map[string]model.Task, edges map[string][]string, visited map[string]bool) {
	if visited[nodeID] {
		return
	}

	node, exists := nodesMap[nodeID]
	if !exists {
		return
	}
	visited[nodeID] = true

	for _, prevNodeID := range node.PrevNodeIDs {
		prevNode, exists := nodesMap[prevNodeID]
		if !exists {
			continue
		}

		shouldProcess := false

		if s.isProxyNode(prevNode) {
			if len(prevNode.PrevNodeIDs) > 0 {
				realPrevID := prevNode.PrevNodeIDs[0]
				// 穿透连接：RealPrev -> Current
				if !slice.Contains(edges[realPrevID], nodeID) {
					edges[realPrevID] = append(edges[realPrevID], nodeID)
				}
				s.processNode(realPrevID, nodesMap, recordMap, edges, visited)
			}
			continue
		}

		switch prevNode.NodeType {
		case model.TaskNode:
			if task, ok := recordMap[prevNodeID]; ok && task.IsFinished == 1 {
				shouldProcess = true
			}
		case model.GateWayNode, model.RootNode:
			shouldProcess = true
		}

		if shouldProcess {
			// 记录边：前置 -> 当前
			if !slice.Contains(edges[prevNodeID], nodeID) {
				edges[prevNodeID] = append(edges[prevNodeID], nodeID)
			}
			s.processNode(prevNodeID, nodesMap, recordMap, edges, visited)
		}
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
