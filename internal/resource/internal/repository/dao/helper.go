package dao

import (
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// buildUpdateDoc 统一合并属性和 utime，消减多处循环拷贝代码，优化内存配给
func (dao *resourceDAO) buildUpdateDoc(data mongox.MapStr, utime int64) bson.M {
	updateDoc := bson.M{
		"utime": utime,
	}
	for k, v := range data {
		updateDoc[k] = v
	}
	return updateDoc
}

// buildExcludeAndFilterBson 统一构建排除 ID 并执行字段过滤的 BSON 条件，消除逻辑重复
func (dao *resourceDAO) buildExcludeAndFilterBson(modelUid string, ids []int64, filter domain.Condition) bson.M {
	filters := bson.M{"model_uid": modelUid}
	if len(ids) > 0 {
		filters["id"] = bson.M{
			"$nin": ids,
		}
	}

	switch filter.Condition {
	case "not_equal":
		filters[filter.Name] = bson.M{"$ne": filter.Input}
	case "equal":
		filters[filter.Name] = filter.Input
	case "contains":
		filters[filter.Name] = bson.M{"$regex": primitive.Regex{Pattern: filter.Input, Options: "i"}}
	}
	return filters
}

// combineFilters 合并基础模型条件与多组 AND/OR 过滤条件，避免查询树组装逻辑冗余
func (dao *resourceDAO) combineFilters(baseFilter bson.M, orConditions []bson.M) interface{} {
	if len(orConditions) == 0 {
		return baseFilter
	}
	if len(orConditions) == 1 {
		return bson.M{
			"$and": []bson.M{
				baseFilter,
				orConditions[0],
			},
		}
	}
	return bson.M{
		"$and": []bson.M{
			baseFilter,
			{"$or": orConditions},
		},
	}
}

func buildBsonCondition(f domain.FilterCondition) bson.M {
	key := f.FieldUID
	val := f.Value

	switch f.Operator {
	case "eq":
		return bson.M{key: val}
	case "ne":
		return bson.M{key: bson.M{"$ne": val}}
	case "contains":
		s, ok := val.(string)
		if !ok {
			return nil
		}
		return bson.M{key: bson.M{"$regex": primitive.Regex{Pattern: s, Options: "i"}}}
	case "gt":
		return bson.M{key: bson.M{"$gt": val}}
	case "lt":
		return bson.M{key: bson.M{"$lt": val}}
	case "gte":
		return bson.M{key: bson.M{"$gte": val}}
	case "lte":
		return bson.M{key: bson.M{"$lte": val}}
	case "in":
		return bson.M{key: bson.M{"$in": val}}
	case "nin":
		return bson.M{key: bson.M{"$nin": val}}
	default:
		return bson.M{key: val}
	}
}

func buildProjection(fields []string) map[string]int {
	// NOTE: 借助 lo.Associate 简化投影初始化，消除显式循环
	projection := lo.Associate(fields, func(v string) (string, int) {
		return v, 1
	})
	projection["_id"] = 0
	projection["id"] = 1
	projection["model_uid"] = 1
	projection["ctime"] = 1
	projection["utime"] = 1
	return projection
}

// buildFilterConditions 解析并组合多组 FilterGroup 过滤规则为 BSON AND/OR 条件树，消灭多处冗余的 Bson 循环组装
func buildFilterConditions(filterGroups []domain.FilterGroup) []bson.M {
	return lo.FilterMap(filterGroups, func(group domain.FilterGroup, _ int) (bson.M, bool) {
		if len(group.Filters) == 0 {
			return nil, false
		}

		andConditions := lo.FilterMap(group.Filters, func(f domain.FilterCondition, _ int) (bson.M, bool) {
			cond := buildBsonCondition(f)
			return cond, cond != nil
		})

		if len(andConditions) == 0 {
			return nil, false
		}
		if len(andConditions) == 1 {
			return andConditions[0], true
		}
		return bson.M{"$and": andConditions}, true
	})
}
