package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type RelationResourceDAO interface {
	CreateResourceRelation(ctx context.Context, mg ResourceRelation) (int64, error)
	ListResourceRelation(ctx context.Context, offset, limit int64) ([]*ResourceRelation, error)

	CountByModelUid(ctx context.Context, modelUid string) (int64, error)

	ListSrcResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error)
	ListDstResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error)

	ListSrcResources(ctx context.Context, modelUid string, id int64) ([]*ResourceRelation, error)
	ListDstResources(ctx context.Context, modelUid string, id int64) ([]*ResourceRelation, error)

	ListSrcAggregated(ctx context.Context, modelUid string, id int64) ([]ResourceAggregatedData, error)
	ListDstAggregated(ctx context.Context, modelUid string, id int64) ([]ResourceAggregatedData, error)
}

func NewRelationResourceDAO(client *mongo.Client) RelationResourceDAO {
	return &resourceDAO{
		db: mongox.NewMongo(client),
	}
}

type resourceDAO struct {
	db *mongox.Mongo
}

func (dao *resourceDAO) ListSrcResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error) {
	col := dao.db.Collection(ResourceRelationCollection)
	filer := bson.M{
		"$and": []bson.M{
			{"source_model_uid": modelUid},
			{"relation_type_uid": relationType},
		},
	}
	opt := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	resp, err := col.Find(ctx, filer, opt)

	var set []int64
	for resp.Next(ctx) {
		var result struct {
			Id int64 `bson:"id"`
		}

		if err = resp.Decode(&result); err != nil {
			return nil, err
		}
		set = append(set, result.Id)
	}

	return set, nil
}

func (dao *resourceDAO) ListDstResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error) {
	col := dao.db.Collection(ResourceRelationCollection)
	filer := bson.M{
		"$and": []bson.M{
			{"target_resource_id": modelUid},
			{"relation_type_uid": relationType},
		},
	}
	opt := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	resp, err := col.Find(ctx, filer, opt)

	var set []int64
	for resp.Next(ctx) {
		var result struct {
			Id int64 `bson:"id"`
		}

		if err = resp.Decode(&result); err != nil {
			return nil, err
		}
		set = append(set, result.Id)
	}

	return set, nil
}

func (dao *resourceDAO) CreateResourceRelation(ctx context.Context, mr ResourceRelation) (int64, error) {
	now := time.Now()
	mr.Ctime, mr.Utime = now.UnixMilli(), now.UnixMilli()
	mr.Id = dao.db.GetIdGenerator(ResourceRelationCollection)
	col := dao.db.Collection(ResourceRelationCollection)

	mr.RelationName = fmt.Sprintf("%s_%s_%s",
		mr.SourceModelUID, mr.RelationTypeUID, mr.TargetModelUID)

	_, err := col.InsertMany(ctx, []interface{}{mr})

	if err != nil {
		return 0, err
	}

	return mr.Id, nil
}

func (dao *resourceDAO) ListResourceRelation(ctx context.Context, offset, limit int64) ([]*ResourceRelation, error) {
	col := dao.db.Collection(ResourceRelationCollection)

	filer := bson.M{}
	opt := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	resp, err := col.Find(ctx, filer, opt)
	var set []*ResourceRelation
	for resp.Next(ctx) {
		ins := &ResourceRelation{}
		if err = resp.Decode(ins); err != nil {
			return nil, err
		}
		set = append(set, ins)
	}

	return set, nil
}

func (dao *resourceDAO) CountByModelUid(ctx context.Context, modelUid string) (int64, error) {
	col := dao.db.Collection(ModelRelationCollection)
	filer := bson.M{"relation_name": bson.M{"$regex": primitive.Regex{Pattern: modelUid, Options: "i"}}}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (dao *resourceDAO) ListResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error) {
	col := dao.db.Collection(ResourceRelationCollection)
	filer := bson.M{
		"$and": []bson.M{
			{"relation_name": bson.M{"$regex": primitive.Regex{Pattern: modelUid, Options: "i"}}},
			{"relation_type_uid": relationType},
		},
	}
	opt := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	resp, err := col.Find(ctx, filer, opt)

	var set []int64
	for resp.Next(ctx) {
		var result struct {
			Id int64 `bson:"id"`
		}

		if err = resp.Decode(&result); err != nil {
			return nil, err
		}
		set = append(set, result.Id)
	}

	return set, nil
}

func (dao *resourceDAO) ListSrcResources(ctx context.Context, modelUid string, id int64) ([]*ResourceRelation, error) {
	col := dao.db.Collection(ResourceRelationCollection)
	filter := bson.M{
		"$and": []bson.M{
			{"source_model_uid": modelUid},
			{"source_resource_id": id},
		},
	}

	opt := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	resp, err := col.Find(ctx, filter, opt)
	var set []*ResourceRelation
	for resp.Next(ctx) {
		ins := &ResourceRelation{}
		if err = resp.Decode(ins); err != nil {
			return nil, err
		}
		set = append(set, ins)
	}

	return set, nil
}

func (dao *resourceDAO) ListSrcAggregated(ctx context.Context, modelUid string, id int64) ([]ResourceAggregatedData, error) {
	col := dao.db.Collection(ResourceRelationCollection)
	filter := bson.M{
		"$and": []bson.M{
			{"source_model_uid": modelUid},
			{"source_resource_id": id},
		},
	}

	pipeline := mongo.Pipeline{
		{{"$match", filter}},
		{{"$group", bson.D{
			{"_id", "$relation_name"},
			{"count", bson.D{{"$sum", 1}}},                         // 统计每个分组中的文档数量
			{"data", bson.D{{"$push", "$$ROOT"}}},                  // 将每个文档添加到一个数组中
			{"model_uid", bson.D{{"$first", "$target_model_uid"}}}, // 添加额外字段
		}}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}

	// 遍历游标，解码每个文档
	var result []ResourceAggregatedData
	for cursor.Next(context.Background()) {

		var rad ResourceAggregatedData
		if err = cursor.Decode(&rad); err != nil {
			return nil, err
		}

		result = append(result, rad)
	}

	return result, nil
}

func (dao *resourceDAO) ListDstAggregated(ctx context.Context, modelUid string, id int64) ([]ResourceAggregatedData, error) {
	col := dao.db.Collection(ResourceRelationCollection)
	filter := bson.M{
		"$and": []bson.M{
			{"target_model_uid": modelUid},
			{"target_resource_id": id},
		},
	}

	pipeline := mongo.Pipeline{
		{{"$match", filter}}, // 添加筛选条件
		{{"$group", bson.D{
			{"_id", "$relation_name"},
			{"count", bson.D{{"$sum", 1}}},                         // 统计每个分组中的文档数量
			{"data", bson.D{{"$push", "$$ROOT"}}},                  // 将每个文档添加到一个数组中
			{"model_uid", bson.D{{"$first", "$source_model_uid"}}}, // 添加额外字段
		}}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}

	// 遍历游标，解码每个文档
	var result []ResourceAggregatedData
	for cursor.Next(context.Background()) {

		var rad ResourceAggregatedData
		if err = cursor.Decode(&rad); err != nil {
			return nil, err
		}

		result = append(result, rad)
	}

	return result, nil
}

func (dao *resourceDAO) ListDstResources(ctx context.Context, modelUid string, id int64) ([]*ResourceRelation, error) {
	col := dao.db.Collection(ResourceRelationCollection)
	filter := bson.M{
		"$and": []bson.M{
			{"target_model_uid": modelUid},
			{"target_resource_id": id},
		},
	}

	opt := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	resp, err := col.Find(ctx, filter, opt)
	var set []*ResourceRelation
	for resp.Next(ctx) {
		ins := &ResourceRelation{}
		if err = resp.Decode(ins); err != nil {
			return nil, err
		}
		set = append(set, ins)
	}

	return set, nil
}

type ResourceRelation struct {
	Id               int64  `bson:"id"`
	SourceModelUID   string `bson:"source_model_uid"`
	TargetModelUID   string `bson:"target_model_uid"`
	SourceResourceID int64  `bson:"source_resource_id"`
	TargetResourceID int64  `bson:"target_resource_id"`
	RelationTypeUID  string `bson:"relation_type_uid"`
	RelationName     string `bson:"relation_name"` // 唯一标识、以防重复创建
	Ctime            int64  `bson:"ctime"`
	Utime            int64  `bson:"utime"`
}

type ResourceAggregatedData struct {
	RelationName string             `bson:"_id"`
	ModelUid     string             `bson:"model_uid"`
	Count        int                `bson:"count"`
	Data         []ResourceRelation `bson:"data"`
}
