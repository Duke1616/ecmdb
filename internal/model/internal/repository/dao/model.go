package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	ModelCollection      = "c_model"
	ModelGroupCollection = "c_model_group"
)

type ModelDAO interface {
	CreateModelGroup(ctx context.Context, mg ModelGroup) (int64, error)
	CreateModel(ctx context.Context, m Model) (int64, error)
	GetModelByUid(ctx context.Context, uid string) (Model, error)
	ListModels(ctx context.Context, offset, limit int64) ([]Model, error)
	CountModels(ctx context.Context) (int64, error)
}

func NewModelDAO(client *mongo.Client) ModelDAO {
	return &modelDAO{
		db: mongox.NewMongo(client),
	}
}

type modelDAO struct {
	db *mongox.Mongo
}

func (dao *modelDAO) CreateModelGroup(ctx context.Context, mg ModelGroup) (int64, error) {
	now := time.Now()
	mg.Ctime, mg.Utime = now.UnixMilli(), now.UnixMilli()
	mg.Id = dao.db.GetIdGenerator(ModelGroupCollection)
	col := dao.db.Collection(ModelGroupCollection)

	_, err := col.InsertMany(ctx, []interface{}{mg})

	if err != nil {
		return 0, err
	}

	return mg.Id, nil
}

func (dao *modelDAO) CreateModel(ctx context.Context, md Model) (int64, error) {
	now := time.Now()
	md.Ctime, md.Utime = now.UnixMilli(), now.UnixMilli()
	md.Id = dao.db.GetIdGenerator(ModelCollection)
	col := dao.db.Collection(ModelCollection)

	_, err := col.InsertMany(ctx, []interface{}{md})

	if err != nil {
		return 0, err
	}

	return md.Id, nil
}

func (dao *modelDAO) GetModelByUid(ctx context.Context, uid string) (Model, error) {
	col := dao.db.Collection(ModelCollection)
	filter := bson.M{"uid": uid}

	var m Model
	if err := col.FindOne(ctx, filter).Decode(&m); err != nil {
		return Model{}, err
	}

	return m, nil
}

func (dao *modelDAO) ListModels(ctx context.Context, offset, limit int64) ([]Model, error) {
	col := dao.db.Collection(ModelCollection)

	filer := bson.M{}
	opt := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	resp, err := col.Find(ctx, filer, opt)
	var set []Model
	for resp.Next(ctx) {
		var ins Model
		if err = resp.Decode(&ins); err != nil {
			return nil, err
		}
		set = append(set, ins)
	}

	return set, nil
}

func (dao *modelDAO) CountModels(ctx context.Context) (int64, error) {
	col := dao.db.Collection(ModelCollection)
	filer := bson.M{}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, err
	}

	return count, nil
}

type ModelGroup struct {
	Id    int64  `bson:"id"`
	Name  string `bson:"name"`
	Ctime int64  `bson:"ctime"`
	Utime int64  `bson:"utime"`
}

type Model struct {
	Id           int64  `bson:"id"`
	ModelGroupId int64  `bson:"model_group_id"`
	Name         string `bson:"name"`
	UID          string `bson:"uid"`
	Icon         string `bson:"icon"`
	Ctime        int64  `bson:"ctime"`
	Utime        int64  `bson:"utime"`
}
