package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"time"
)

type ModelGroupDAO interface {
	CreateModelGroup(ctx context.Context, mg ModelGroup) (int64, error)
}

func NewModelGroupDAO(db *mongox.Mongo) ModelGroupDAO {
	return &groupDAO{
		db: db,
	}
}

type groupDAO struct {
	db *mongox.Mongo
}

func (dao *groupDAO) CreateModelGroup(ctx context.Context, mg ModelGroup) (int64, error) {
	session, err := dao.db.DBClient.StartSession()
	if err != nil {
		return 0, fmt.Errorf("无法创建会话: %w", err)
	}
	defer session.EndSession(ctx)

	// 开始事务
	err = session.StartTransaction()
	if err != nil {
		return 0, fmt.Errorf("无法开始事务: %w", err)
	}

	now := time.Now()
	mg.Ctime, mg.Utime = now.UnixMilli(), now.UnixMilli()
	mg.Id = dao.db.GetIdGenerator(ModelGroupCollection)
	col := dao.db.Collection(ModelGroupCollection)

	_, err = col.InsertOne(ctx, mg)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	// 提交事务
	err = session.CommitTransaction(ctx)
	if err != nil {
		return 0, fmt.Errorf("提交事务错误: %w", err)
	}

	return mg.Id, nil
}
