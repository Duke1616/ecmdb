package service

import (
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/workflow/pkg/easyflow"
)

// NodeStatusAnalyzer 负责分析节点和批次的状态
// 封装了对 BatchStats、IsCosigned、SystemPass 的判断逻辑
type NodeStatusAnalyzer struct {
	// Key: NodeID, Value: Map[BatchCode]Stats
	stats    map[string]map[string]*BatchStats
	nodesMap map[string]model.Node
}

type BatchStats struct {
	Total      int
	Passed     int // Status = 1
	Pending    int // IsFinished = 0
	SystemPass int // Status = 3
}

func NewNodeStatusAnalyzer(records []model.Task, nodesMap map[string]model.Node) *NodeStatusAnalyzer {
	stats := make(map[string]map[string]*BatchStats)

	for _, t := range records {
		if _, ok := stats[t.NodeID]; !ok {
			stats[t.NodeID] = make(map[string]*BatchStats)
		}

		batchStats, ok := stats[t.NodeID][t.BatchCode]
		if !ok {
			batchStats = &BatchStats{}
			stats[t.NodeID][t.BatchCode] = batchStats
		}

		batchStats.Total++
		if t.IsFinished == 0 {
			batchStats.Pending++
		} else if t.Status == 1 {
			batchStats.Passed++
		} else if t.Status == 3 {
			batchStats.SystemPass++
		}
	}

	return &NodeStatusAnalyzer{
		stats:    stats,
		nodesMap: nodesMap,
	}
}

// IsBatchTainted 判定当前批次是否由 SystemPass (自动跳过/取消) 污染
// 如果是，通常意味着该分支被废弃，不应绘制入边
func (a *NodeStatusAnalyzer) IsBatchTainted(nodeID, batchCode string) bool {
	if batches, ok := a.stats[nodeID]; ok {
		if stats := batches[batchCode]; stats != nil {
			return stats.SystemPass > 0
		}
	}
	return false
}

// IsBatchEffectivelyPassed 判定当前批次是否"有效通过"
// 必须满足：非会签简单通过 或 会签全部通过
func (a *NodeStatusAnalyzer) IsBatchEffectivelyPassed(task model.Task) bool {
	// 基础检查：必须是 finish 且 status=1 (pass)
	if task.IsFinished != 1 || task.Status != 1 {
		return false
	}

	// 检查会签
	isCosigned := false
	if nodeDef, exists := a.nodesMap[task.NodeID]; exists && nodeDef.IsCosigned == 1 {
		isCosigned = true
	} else if task.IsCosigned == 1 {
		isCosigned = true
	}

	if isCosigned {
		if batches, ok := a.stats[task.NodeID]; ok {
			stats := batches[task.BatchCode]
			// 会签要求：所有任务都必须是 Status 1 (Passed)
			if stats != nil && stats.Passed < stats.Total {
				return false
			}
		}
	}

	return true
}

// HasNewerBatchPending 检查是否存在更新的批次正在进行中 (Loop-back)
// 如果有，说明流程回退到了这里，旧的批次路径应该被阻塞
func (a *NodeStatusAnalyzer) HasNewerBatchPending(nodeID, currentBatchCode string) bool {
	batches, ok := a.stats[nodeID]
	if !ok {
		return false
	}

	for batchCode, stats := range batches {
		if batchCode != currentBatchCode && stats.Pending > 0 {
			return true
		}
	}
	return false
}

// GraphTopologyService 负责处理图结构和路径查找
// 屏蔽 Proxy 穿透、逻辑前置查找等复杂性
type GraphTopologyService struct {
	nodesMap       map[string]model.Node
	effectiveGraph map[string][]string // 剔除 Proxy 后的邻接表 Graph[Src] -> []Dst
}

func NewGraphTopologyService(nodesMap map[string]model.Node, svc *service) *GraphTopologyService {
	return &GraphTopologyService{
		nodesMap:       nodesMap,
		effectiveGraph: svc.buildEffectiveGraph(nodesMap),
	}
}

// ResolveLogicalPrev 解析逻辑前置节点 (穿透 Proxy)
func (g *GraphTopologyService) ResolveLogicalPrev(rawPrevID string) string {
	node, ok := g.nodesMap[rawPrevID]
	if !ok {
		return rawPrevID
	}

	if g.isProxyNode(node) {
		if len(node.PrevNodeIDs) > 0 {
			return g.ResolveLogicalPrev(node.PrevNodeIDs[0])
		}
	}
	return rawPrevID
}

func (g *GraphTopologyService) isProxyNode(node model.Node) bool {
	for _, uid := range node.UserIDs {
		if uid == easyflow.SysAutoUser {
			return true
		}
	}
	return false
}

// FindPath 寻找两个节点之间的路径 (穿透网关)
func (g *GraphTopologyService) FindPath(startID, endID string, svc *service) []string {
	return svc.findPathThroughGateways(startID, endID, g.effectiveGraph, g.nodesMap)
}

// ResolveNextNodes 获取当前节点的后续节点 (用于前向 Look-ahead)
func (g *GraphTopologyService) ResolveNextNodes(nodeID string) []string {
	return g.effectiveGraph[nodeID]
}
