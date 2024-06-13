package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/xen0n/go-workwx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// ErrDataNotFound 通用的数据没找到
var ErrDataNotFound = mongo.ErrNoDocuments

const (
	TemplateCollection = "c_template"
)

type TemplateDAO interface {
	CreateTemplate(ctx context.Context, t Template) (int64, error)
	FindByHash(ctx context.Context, hash string) (Template, error)
	DetailTemplate(ctx context.Context, id int64) (Template, error)
	ListTemplate(ctx context.Context, offset, limit int64) ([]Template, error)
	Count(ctx context.Context) (int64, error)
}

func NewTemplateDAO(db *mongox.Mongo) TemplateDAO {
	return &templateDAO{
		db: db,
	}
}

type templateDAO struct {
	db *mongox.Mongo
}

func (dao *templateDAO) CreateTemplate(ctx context.Context, t Template) (int64, error) {
	t.Id = dao.db.GetIdGenerator(TemplateCollection)
	col := dao.db.Collection(TemplateCollection)
	now := time.Now()
	t.Ctime, t.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, t)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return t.Id, nil
}

func (dao *templateDAO) FindByHash(ctx context.Context, hash string) (Template, error) {
	col := dao.db.Collection(TemplateCollection)
	var t Template
	filter := bson.M{"unique_hash": hash}

	if err := col.FindOne(ctx, filter).Decode(&t); err != nil {
		return Template{}, fmt.Errorf("解码错误，%w", err)
	}

	return t, nil
}

func (dao *templateDAO) DetailTemplate(ctx context.Context, id int64) (Template, error) {
	col := dao.db.Collection(TemplateCollection)
	filter := bson.M{"id": id}

	var t Template
	if err := col.FindOne(ctx, filter).Decode(&t); err != nil {
		return Template{}, fmt.Errorf("解码错误，%w", err)
	}

	return t, nil
}

func (dao *templateDAO) ListTemplate(ctx context.Context, offset, limit int64) ([]Template, error) {
	col := dao.db.Collection(TemplateCollection)
	filter := bson.M{}
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	cursor, err := col.Find(ctx, filter, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []Template
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *templateDAO) Count(ctx context.Context) (int64, error) {
	col := dao.db.Collection(TemplateCollection)
	filer := bson.M{}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

type Template struct {
	Id               int64                     `bson:"id"`
	Name             string                    `bson:"name"`
	CreateType       uint8                     `bson:"create_type"`
	Rules            []map[string]interface{}  `bson:"rules"`
	Options          map[string]interface{}    `bson:"options"`
	WechatOAControls workwx.OATemplateControls `bson:"wechat_oa_controls,omitempty"`
	UniqueHash       string                    `bson:"unique_hash"`
	Desc             string                    `bson:"desc,omitempty"`
	Ctime            int64                     `bson:"ctime"`
	Utime            int64                     `bson:"utime"`
}
