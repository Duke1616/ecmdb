package web

import "github.com/Duke1616/ecmdb/pkg/mongox"

// Resource 单个资产对象展示视图
type Resource struct {
	ID       int64         `json:"id"`
	Name     string        `json:"name"`
	ModelUID string        `json:"model_uid"`
	Data     mongox.MapStr `json:"data"`
}

// RetrieveResources 资产列表查询结果
type RetrieveResources struct {
	Resources []Resource `json:"resources"`
	Total     int64      `json:"total"`
}

// RetrieveSearchResources 全文检索结果视图
type RetrieveSearchResources struct {
	ModelUid string          `json:"model_uid"`
	Total    int             `json:"total"`
	Data     []mongox.MapStr `json:"data"`
}

// RetrieveAggregatedAssets 资产关联关系聚合统计视图
type RetrieveAggregatedAssets struct {
	RelationName string  `json:"relation_name"`
	ModelUid     string  `json:"model_uid"`
	Total        int     `json:"total"`
	ResourceIds  []int64 `json:"resource_ids"`
}

// RetrieveGraph 资产关联拓扑图全貌视图
type RetrieveGraph struct {
	RootId string       `json:"rootId"`
	Nodes  []Node       `json:"nodes"`
	Lines  []Line       `json:"lines"`
	Models []GraphModel `json:"models,omitempty"`
}

// GraphModel 拓扑图模型图例元数据
type GraphModel struct {
	ModelUID  string `json:"model_uid"`
	ModelName string `json:"model_name"`
	Icon      string `json:"icon,omitempty"`
}

// Node 拓扑图单个节点
type Node struct {
	ID                   string         `json:"id"`
	Text                 string         `json:"text"`
	ExpandHolderPosition string         `json:"expandHolderPosition,omitempty"`
	Expanded             bool           `json:"expanded"`
	Data                 map[string]any `json:"data,omitempty"`
}

// Line 拓扑图连接边
type Line struct {
	From string `json:"from"`
	To   string `json:"to"`
}
