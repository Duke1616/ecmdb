package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type RelationTypeDAO interface {
	Create(ctx context.Context, r RelationType) (int64, error)
}

func NewRelationTypeDAO(client *mongo.Client) RelationTypeDAO {
	return &relationDAO{
		db: mongox.NewMongo(client),
	}
}

type relationDAO struct {
	db *mongox.Mongo
}

func (dao *relationDAO) Create(ctx context.Context, r RelationType) (int64, error) {
	now := time.Now()
	r.Ctime, r.Utime = now.UnixMilli(), now.UnixMilli()
	r.Id = dao.db.GetIdGenerator(RelationTypeCollection)
	col := dao.db.Collection(RelationTypeCollection)

	_, err := col.InsertMany(ctx, []interface{}{r})

	if err != nil {
		return 0, err
	}

	return r.Id, nil
}

type RelationType struct {
	Id             int64  `bson:"id"`
	UID            string `bson:"uid"`
	SourceDescribe string `bson:"source_describe"`
	TargetDescribe string `bson:"target_describe"`
	Ctime          int64  `bson:"ctime"`
	Utime          int64  `bson:"utime"`
}
