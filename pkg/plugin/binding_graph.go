package plugin

import (
	"fmt"
	"slices"
	"strings"

	"github.com/samber/lo"
)

type mutableBindingGraphIndex struct {
	nodesByID    map[string]*BindingGraphNode
	childrenByID map[string][]*BindingGraphEdge
}

type compiledBindingGraphIndex struct {
	entryNode     BindingGraphNode
	nodesByID     map[string]BindingGraphNode
	childrenByID  map[string][]BindingGraphEdge
	parentByChild map[string]BindingGraphEdge
}

func GraphEntryNode(graph *BindingGraph) (BindingGraphNode, bool) {
	if graph == nil {
		return BindingGraphNode{}, false
	}
	return lo.Find(graph.Nodes, func(node BindingGraphNode) bool {
		return node.ID == graph.EntryNodeID
	})
}

func BuildCenterGraph[T any](name string, modelUID string) (*BindingGraph, error) {
	spec, err := BuildCenterSpec[T](name, modelUID)
	if err != nil {
		return nil, err
	}
	return GraphFromBindingSpecs(modelUID, []ResourceSpec{spec})
}

func MutateBindingGraphPath(
	graph *BindingGraph,
	path string,
	fn func(node *BindingGraphNode, edge *BindingGraphEdge),
) bool {
	if graph == nil || fn == nil {
		return false
	}
	parts := splitBindingGraphPath(path)
	if len(parts) == 0 {
		return false
	}
	index := indexMutableBindingGraph(graph)
	node, edge, ok := index.resolvePath(graph.EntryNodeID, parts)
	if !ok {
		return false
	}
	if len(parts) == 1 {
		fn(node, nil)
		return true
	}
	fn(node, edge)
	return true
}

func splitBindingGraphPath(path string) []string {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}

	parts := lo.Map(strings.Split(path, "."), func(part string, _ int) string {
		return strings.TrimSpace(part)
	})
	if lo.Contains(parts, "") {
		return nil
	}
	return parts
}

func indexMutableBindingGraph(graph *BindingGraph) mutableBindingGraphIndex {
	index := mutableBindingGraphIndex{
		nodesByID:    make(map[string]*BindingGraphNode, len(graph.Nodes)),
		childrenByID: make(map[string][]*BindingGraphEdge),
	}
	for i := range graph.Nodes {
		node := &graph.Nodes[i]
		index.nodesByID[node.ID] = node
	}
	for i := range graph.Edges {
		edge := &graph.Edges[i]
		index.childrenByID[edge.From] = append(index.childrenByID[edge.From], edge)
	}
	return index
}

func (index mutableBindingGraphIndex) resolvePath(
	entryNodeID string,
	parts []string,
) (*BindingGraphNode, *BindingGraphEdge, bool) {
	current, ok := index.nodesByID[entryNodeID]
	if !ok || current.Name != parts[0] {
		return nil, nil, false
	}
	if len(parts) == 1 {
		return current, nil, true
	}

	var incoming *BindingGraphEdge
	for _, part := range parts[1:] {
		nextNode, nextIncoming, ok := index.childByName(current.ID, part)
		if !ok {
			return nil, nil, false
		}
		current = nextNode
		incoming = nextIncoming
	}
	return current, incoming, true
}

func (index mutableBindingGraphIndex) childByName(parentID string, name string) (*BindingGraphNode, *BindingGraphEdge, bool) {
	edge, ok := lo.Find(index.childrenByID[parentID], func(edge *BindingGraphEdge) bool {
		node := index.nodesByID[edge.To]
		return node != nil && node.Name == name
	})
	if !ok {
		return nil, nil, false
	}
	return index.nodesByID[edge.To], edge, true
}

func CompileBindingGraph(graph *BindingGraph) ([]ResourceSpec, error) {
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
	return []ResourceSpec{root}, nil
}

func GraphFromBindingSpecs(modelUID string, specs []ResourceSpec) (*BindingGraph, error) {
	if len(specs) == 0 {
		return nil, nil
	}

	root := graphRootSpec(modelUID, specs)
	entryNodeID := root.Name
	if entryNodeID == "" {
		entryNodeID = "root"
	}
	graph := &BindingGraph{
		EntryNodeID: entryNodeID,
	}

	appendSpecToGraph(graph, root, graph.EntryNodeID)
	return graph, nil
}

func indexCompiledBindingGraph(graph *BindingGraph) (compiledBindingGraphIndex, error) {
	index := compiledBindingGraphIndex{
		nodesByID:     make(map[string]BindingGraphNode, len(graph.Nodes)),
		childrenByID:  make(map[string][]BindingGraphEdge),
		parentByChild: make(map[string]BindingGraphEdge, len(graph.Edges)),
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
	incoming *BindingGraphEdge,
	visiting map[string]bool,
	visited map[string]bool,
) (ResourceSpec, error) {
	if visiting[nodeID] {
		return ResourceSpec{}, fmt.Errorf("graph 存在循环依赖: %s", nodeID)
	}
	if visited[nodeID] {
		return ResourceSpec{}, fmt.Errorf("graph 节点重复访问: %s", nodeID)
	}

	visiting[nodeID] = true
	spec, err := newResourceSpecFromGraphNode(index.nodesByID[nodeID], nodeID, incoming)
	if err != nil {
		return ResourceSpec{}, err
	}

	for _, edge := range index.childrenByID[nodeID] {
		child, err := index.compileNode(edge.To, &edge, visiting, visited)
		if err != nil {
			return ResourceSpec{}, err
		}
		spec.Children = append(spec.Children, child)
	}

	delete(visiting, nodeID)
	visited[nodeID] = true
	return spec, nil
}

func newResourceSpecFromGraphNode(
	node BindingGraphNode,
	nodeID string,
	incoming *BindingGraphEdge,
) (ResourceSpec, error) {
	spec := ResourceSpec{
		Name:        node.Name,
		ModelUID:    node.ModelUID,
		Cardinality: defaultCardinality(node.Cardinality),
		Required:    node.Required,
		Fields:      make(map[string]string, len(node.FieldMappings)),
		Filters:     slices.Clone(node.Filters),
	}
	for _, mapping := range node.FieldMappings {
		if mapping.Input == "" {
			return ResourceSpec{}, fmt.Errorf("graph 节点字段 input 不能为空: %s", nodeID)
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

func graphRootSpec(modelUID string, specs []ResourceSpec) ResourceSpec {
	if spec, ok := lo.Find(specs, func(spec ResourceSpec) bool {
		if spec.ModelUID == modelUID && spec.RelationType == "" {
			return true
		}
		return false
	}); ok {
		return spec
	}
	return specs[0]
}

func appendSpecToGraph(graph *BindingGraph, spec ResourceSpec, nodeID string) {
	graph.Nodes = append(graph.Nodes, BindingGraphNode{
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
		graph.Edges = append(graph.Edges, BindingGraphEdge{
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
		return CardinalityOne
	}
	return value
}

func mappingsFromSpec(spec ResourceSpec) []FieldMapping {
	if len(spec.Fields) == 0 {
		return nil
	}
	required := make(map[string]struct{}, len(spec.RequiredFields))
	for _, field := range spec.RequiredFields {
		required[field] = struct{}{}
	}

	keys := make([]string, 0, len(spec.Fields))
	for input := range spec.Fields {
		keys = append(keys, input)
	}
	slices.Sort(keys)

	mappings := make([]FieldMapping, 0, len(keys))
	for _, input := range keys {
		_, ok := required[input]
		mappings = append(mappings, FieldMapping{
			Input:         input,
			ResourceField: spec.Fields[input],
			Required:      ok,
		})
	}
	return mappings
}
