package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type RelationTypeDAO interface {
	Create(ctx context.Context, r RelationType) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]*RelationType, error)
	Count(ctx context.Context) (int64, error)
}

func NewRelationTypeDAO(db *mongox.Mongo) RelationTypeDAO {
	return &relationDAO{
		db: db,
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

func (dao *relationDAO) List(ctx context.Context, offset, limit int64) ([]*RelationType, error) {
	col := dao.db.Collection(RelationTypeCollection)

	filer := bson.M{}
	opt := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	resp, err := col.Find(ctx, filer, opt)
	var set []*RelationType
	for resp.Next(ctx) {
		ins := &RelationType{}
		if err = resp.Decode(ins); err != nil {
			return nil, err
		}
		set = append(set, ins)
	}

	return set, nil
}

func (dao *relationDAO) Count(ctx context.Context) (int64, error) {
	col := dao.db.Collection(RelationTypeCollection)
	filer := bson.M{}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, err
	}

	return count, nil
}

type RelationType struct {
	Id             int64  `bson:"id"`
	Name           string `bson:"name"`
	UID            string `bson:"uid"`
	SourceDescribe string `bson:"source_describe"`
	TargetDescribe string `bson:"target_describe"`
	Ctime          int64  `bson:"ctime"`
	Utime          int64  `bson:"utime"`
}
