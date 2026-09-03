package mongox

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

// ErrMissingTenantContext 多租户安全拦截：未显式声明 IgnoreTenant 且缺失有效租户上下文
var ErrMissingTenantContext = errors.New("多租户安全拦截：未显式声明 IgnoreTenant 且缺失有效租户上下文，请使用 mongox.IgnoreTenantContext(ctx) 显式提权")

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

// IsNotFoundError 检查是否是数据不存在错误
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// 判断是否是 mongo.ErrNoDocuments
	return errors.Is(err, mongo.ErrNoDocuments)
}
