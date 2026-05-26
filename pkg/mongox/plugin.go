package mongox

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

// Statement 承载 MongoDB CRUD 操作生命周期的运行期元数据
type Statement struct {
	CollectionName string
	Model          interface{} // 单个实体的指针，或者切片指针（例如 &User 或 &[]User）
	Filter         bson.M      // 最终传递给底层驱动的查询过滤器
	Update         interface{} // 更新指令（如 bson.M{"$set": ...}）
	Context        context.Context
}

// IMongoPlugin 定义多插件拦截规范名
type IMongoPlugin interface {
	Name() string
}

// BeforeFinder 拦截查询前置节点（FindOne, Find, CountDocuments, Distinct 等）
type BeforeFinder interface {
	BeforeFind(stmt *Statement) error
}

// BeforeInserter 拦截写入前置节点（InsertOne, InsertMany 等）
type BeforeInserter interface {
	BeforeInsert(stmt *Statement) error
}

// BeforeUpdater 拦截更新前置节点（UpdateOne, UpdateMany, ReplaceOne 等）
type BeforeUpdater interface {
	BeforeUpdate(stmt *Statement) error
}

// BeforeDeleter 拦截删除前置节点（DeleteOne, DeleteMany 等）
type BeforeDeleter interface {
	BeforeDelete(stmt *Statement) error
}
