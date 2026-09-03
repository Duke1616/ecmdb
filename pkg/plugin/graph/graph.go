package graph

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
)

// GraphEntryNode 获取绑定图的入口节点
func GraphEntryNode(graph *types.BindingGraph) (types.BindingGraphNode, bool) {
	if graph == nil {
		return types.BindingGraphNode{}, false
	}
	return lo.Find(graph.Nodes, func(node types.BindingGraphNode) bool {
		return node.ID == graph.EntryNodeID
	})
}

type compiledBindingGraphIndex struct {
	entryNode     types.BindingGraphNode
	nodesByID     map[string]types.BindingGraphNode
	childrenByID  map[string][]types.BindingGraphEdge
	parentByChild map[string]types.BindingGraphEdge
}

// CompileBindingGraph 校验图的合法性（单入口、无环连通树）并编译成层级 ResourceSpec 树
func CompileBindingGraph(graph *types.BindingGraph) ([]types.ResourceSpec, error) {
	if graph == nil || len(graph.Nodes) == 0 {
		return nil, nil
	}
	index, err := indexCompiledBindingGraph(graph)
	if err != nil {
		return nil, err
	}

	visiting := make(map[string]bool, len(index.nodesByID))
	visited := make(map[string]bool, len(index.nodesByID))

	root, err := index.compileNode(index.entryNode.ID, nil, visiting, visited)
	if err != nil {
		return nil, err
	}
	if len(visited) != len(index.nodesByID) {
		return nil, fmt.Errorf("graph 存在未连接到入口的节点")
	}
	return []types.ResourceSpec{root}, nil
}

// GraphFromBindingSpecs 将声明式的 ResourceSpec 列表转化为 BindingGraph 存储模型
func GraphFromBindingSpecs(modelUID string, specs []types.ResourceSpec) (*types.BindingGraph, error) {
	if len(specs) == 0 {
		return nil, nil
	}

	root := graphRootSpec(modelUID, specs)
	entryNodeID := root.Name
	if entryNodeID == "" {
		entryNodeID = "root"
	}
	graph := &types.BindingGraph{
		EntryNodeID: entryNodeID,
	}

	appendSpecToGraph(graph, root, graph.EntryNodeID)
	return graph, nil
}

func indexCompiledBindingGraph(graph *types.BindingGraph) (compiledBindingGraphIndex, error) {
	index := compiledBindingGraphIndex{
		nodesByID:     make(map[string]types.BindingGraphNode, len(graph.Nodes)),
		childrenByID:  make(map[string][]types.BindingGraphEdge),
		parentByChild: make(map[string]types.BindingGraphEdge, len(graph.Edges)),
	}

	for _, node := range graph.Nodes {
		if node.ID == "" {
			return compiledBindingGraphIndex{}, fmt.Errorf("graph node id 不能为空")
		}
		if _, exists := index.nodesByID[node.ID]; exists {
			return compiledBindingGraphIndex{}, fmt.Errorf("graph node 重复: %s", node.ID)
		}
		index.nodesByID[node.ID] = node
	}

	entryNode, ok := index.nodesByID[graph.EntryNodeID]
	if !ok {
		return compiledBindingGraphIndex{}, fmt.Errorf("graph 入口节点不存在: %s", graph.EntryNodeID)
	}
	index.entryNode = entryNode

	for _, edge := range graph.Edges {
		if _, ok := index.nodesByID[edge.From]; !ok {
			return compiledBindingGraphIndex{}, fmt.Errorf("graph edge.from 节点不存在: %s", edge.From)
		}
		if _, ok := index.nodesByID[edge.To]; !ok {
			return compiledBindingGraphIndex{}, fmt.Errorf("graph edge.to 节点不存在: %s", edge.To)
		}
		if _, exists := index.parentByChild[edge.To]; exists {
			return compiledBindingGraphIndex{}, fmt.Errorf("graph 节点存在多个父节点: %s", edge.To)
		}
		index.parentByChild[edge.To] = edge
		index.childrenByID[edge.From] = append(index.childrenByID[edge.From], edge)
	}

	if _, hasParent := index.parentByChild[index.entryNode.ID]; hasParent {
		return compiledBindingGraphIndex{}, fmt.Errorf("graph 入口节点不能有父节点: %s", index.entryNode.ID)
	}
	return index, nil
}

func (index compiledBindingGraphIndex) compileNode(
	nodeID string,
	incoming *types.BindingGraphEdge,
	visiting map[string]bool,
	visited map[string]bool,
) (types.ResourceSpec, error) {
	if visiting[nodeID] {
		return types.ResourceSpec{}, fmt.Errorf("graph 存在循环依赖: %s", nodeID)
	}
	if visited[nodeID] {
		return types.ResourceSpec{}, fmt.Errorf("graph 节点重复访问: %s", nodeID)
	}

	visiting[nodeID] = true
	spec, err := newResourceSpecFromGraphNode(index.nodesByID[nodeID], nodeID, incoming)
	if err != nil {
		return types.ResourceSpec{}, err
	}

	for _, edge := range index.childrenByID[nodeID] {
		child, err := index.compileNode(edge.To, &edge, visiting, visited)
		if err != nil {
			return types.ResourceSpec{}, err
		}
		spec.Children = append(spec.Children, child)
	}

	delete(visiting, nodeID)
	visited[nodeID] = true
	return spec, nil
}

func newResourceSpecFromGraphNode(
	node types.BindingGraphNode,
	nodeID string,
	incoming *types.BindingGraphEdge,
) (types.ResourceSpec, error) {
	spec := types.ResourceSpec{
		Name:        node.Name,
		ModelUID:    node.ModelUID,
		Cardinality: defaultCardinality(node.Cardinality),
		Required:    node.Required,
		Fields:      make(map[string]string, len(node.FieldMappings)),
		Filters:     slices.Clone(node.Filters),
	}
	for _, mapping := range node.FieldMappings {
		if mapping.Input == "" {
			return types.ResourceSpec{}, fmt.Errorf("graph 节点字段 input 不能为空: %s", nodeID)
		}
		spec.Fields[mapping.Input] = mapping.ResourceField
		if mapping.Required {
			spec.RequiredFields = append(spec.RequiredFields, mapping.Input)
		}
	}
	if incoming != nil {
		spec.RelationType = incoming.RelationType
		spec.Direction = incoming.Direction
	}
	return spec, nil
}

func graphRootSpec(modelUID string, specs []types.ResourceSpec) types.ResourceSpec {
	if spec, ok := lo.Find(specs, func(spec types.ResourceSpec) bool {
		return spec.ModelUID == modelUID && spec.RelationType == ""
	}); ok {
		return spec
	}
	return specs[0]
}

func appendSpecToGraph(graph *types.BindingGraph, spec types.ResourceSpec, nodeID string) {
	graph.Nodes = append(graph.Nodes, types.BindingGraphNode{
		ID:            nodeID,
		Name:          spec.Name,
		ModelUID:      spec.ModelUID,
		Cardinality:   defaultCardinality(spec.Cardinality),
		Required:      spec.Required,
		FieldMappings: mappingsFromSpec(spec),
		Filters:       slices.Clone(spec.Filters),
	})

	for index, child := range spec.Children {
		childID := fmt.Sprintf("%s.%d", nodeID, index)
		graph.Edges = append(graph.Edges, types.BindingGraphEdge{
			From:         nodeID,
			To:           childID,
			RelationType: child.RelationType,
			Direction:    child.Direction,
		})
		appendSpecToGraph(graph, child, childID)
	}
}

func defaultCardinality(value string) string {
	if value == "" {
		return types.CardinalityOne
	}
	return value
}

func mappingsFromSpec(spec types.ResourceSpec) []types.FieldMapping {
	if len(spec.Fields) == 0 {
		return nil
	}

	keys := lo.Keys(spec.Fields)
	slices.Sort(keys)

	return lo.Map(keys, func(input string, _ int) types.FieldMapping {
		return types.FieldMapping{
			Input:         input,
			ResourceField: spec.Fields[input],
			Required:      lo.Contains(spec.RequiredFields, input),
		}
	})
}

// PrepareBindings 规范化并校验插件的所有绑定
func PrepareBindings(pluginID string, bindings []types.Binding) ([]types.Binding, error) {
	if len(bindings) == 0 {
		return nil, fmt.Errorf("bindings 不能为空")
	}
	return lo.MapErr(bindings, func(b types.Binding, _ int) (types.Binding, error) {
		b.PluginID = pluginID
		return PrepareBinding(b)
	})
}

// PrepareBinding 规范化入口模型 UID 并校验单条 Binding
func PrepareBinding(binding types.Binding) (types.Binding, error) {
	var err error
	binding, err = NormalizeBindingGraph(binding)
	if err != nil {
		return types.Binding{}, err
	}
	if err = binding.Validate(); err != nil {
		return types.Binding{}, err
	}
	return binding, nil
}

// NormalizeBindingGraph 校验图并根据入口节点校准 ModelUID
func NormalizeBindingGraph(binding types.Binding) (types.Binding, error) {
	if binding.Graph == nil || len(binding.Graph.Nodes) == 0 {
		return binding, nil
	}
	if _, err := CompileBindingGraph(binding.Graph); err != nil {
		return types.Binding{}, err
	}
	if entry, ok := GraphEntryNode(binding.Graph); ok {
		binding.ModelUID = strings.TrimSpace(entry.ModelUID)
	}
	return binding, nil
}
