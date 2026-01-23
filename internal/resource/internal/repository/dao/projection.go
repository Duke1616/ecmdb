package dao

import (
	"go.mongodb.org/mongo-driver/bson"
)

// buildProjection 构建 MongoDB projection,过滤空字符串字段
// NOTE: MongoDB 不允许空字符串作为字段名,会报错 "FieldPath cannot be constructed with empty string"
func buildProjection(fields []string) bson.M {
	projection := make(bson.M, len(fields)+4) // +4 for _id, id, name, model_uid

	// 过滤空字符串字段
	for _, field := range fields {
		if field != "" {
			projection[field] = 1
		}
	}

	// 添加必需字段
	projection["_id"] = 0
	projection["id"] = 1
	projection["name"] = 1
	projection["model_uid"] = 1

	return projection
}
