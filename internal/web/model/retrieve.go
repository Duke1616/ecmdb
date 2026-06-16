package web

import (
	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/samber/lo"
)

type RetrieveModelGroupedListResp struct {
	Total  int64            `json:"total"`
	Groups []ModelGroupItem `json:"groups"`
	Models []ModelSummaryVO `json:"models"`
}

type ModelGroupItem struct {
	GroupID   int64    `json:"group_id"`
	GroupName string   `json:"group_name"`
	ModelUIDs []string `json:"model_uids"`
}

type ModelSummaryVO struct {
	ID            int64  `json:"id"`
	GroupID       int64  `json:"group_id"`
	Name          string `json:"name"`
	UID           string `json:"uid"`
	Icon          string `json:"icon"`
	ResourceCount int    `json:"resource_count"`
	Builtin       bool   `json:"builtin"`
}

func groupModelUIDsByGroupID(models []domain.Model) map[int64][]string {
	modelsByGroupID := lo.GroupBy(models, func(src domain.Model) int64 {
		return src.GroupId
	})

	return lo.MapValues(modelsByGroupID, func(items []domain.Model, _ int64) []string {
		return lo.Map(items, func(src domain.Model, _ int) string {
			return src.UID
		})
	})
}

func retrieveModelGroups(models []domain.Model, modelGroups []domain.ModelGroup) []ModelGroupItem {
	mds := groupModelUIDsByGroupID(models)
	return lo.Map(modelGroups, func(src domain.ModelGroup, idx int) ModelGroupItem {
		modelUIDs, ok := mds[src.ID]
		if !ok {
			modelUIDs = make([]string, 0)
		}

		return ModelGroupItem{
			GroupID:   src.ID,
			GroupName: src.Name,
			ModelUIDs: modelUIDs,
		}
	})
}

func retrieveModelSummaries(models []domain.Model, resourceCount map[string]int) []ModelSummaryVO {
	return lo.Map(models, func(src domain.Model, idx int) ModelSummaryVO {
		return ModelSummaryVO{
			ID:            src.ID,
			GroupID:       src.GroupId,
			Name:          src.Name,
			UID:           src.UID,
			Icon:          src.Icon,
			ResourceCount: resourceCount[src.UID],
			Builtin:       src.Builtin,
		}
	})
}
