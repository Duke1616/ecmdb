package mongox

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

// IsUniqueConstraintError 检查是否是唯一索引冲突错误
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	// 判断是否是 WriteException
	var we mongo.WriteException
	if errors.As(err, &we) {
		for _, e := range we.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
	}

	// BulkWriteException
	var bwe mongo.BulkWriteException
	if errors.As(err, &bwe) {
		for _, e := range bwe.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
	}

	return false
}
