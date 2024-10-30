package web

import (
	"github.com/Duke1616/ecmdb/internal/model/internal/domain"
	"github.com/ecodeclub/ekit/slice"
)

type ModelListByGroupId struct {
	GroupId   int64   `json:"group_id"`
	GroupName string  `json:"group_name"`
	Models    []Model `json:"models"`
}

type RetrieveModelListByGroupId struct {
	Mgs []ModelListByGroupId `json:"mgs"`
}

// 组合前端数据
func groupModelsByGroupId(models []domain.Model, resourceCount map[string]int) map[int64][]Model {
	mds := make(map[int64][]Model)

	for _, m := range models {
		// 计算模型资源数量
		count := resourceCount[m.UID]

		// 构建 Model
		model := Model{
			Id:    m.ID,
			Name:  m.Name,
			Total: count,
			UID:   m.UID,
			Icon:  m.Icon,
		}

		// 按 GroupId 分组
		mds[m.GroupId] = append(mds[m.GroupId], model)
	}

	return mds
}

// 前端展示
func retrieveModelListByGroupId(models []domain.Model, modelGroups []domain.ModelGroup, resourceCount map[string]int) []ModelListByGroupId {
	mds := groupModelsByGroupId(models, resourceCount)
	return slice.Map(modelGroups, func(idx int, src domain.ModelGroup) ModelListByGroupId {
		return ModelListByGroupId{
			GroupId:   src.ID,
			GroupName: src.Name,
			Models:    mds[src.ID],
		}
	})
}
