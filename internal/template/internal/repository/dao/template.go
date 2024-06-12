package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/xen0n/go-workwx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

type Template struct {
	Id               int64                     `bson:"id"`
	Name             string                    `bson:"name"`
	CreateType       uint8                     `bson:"create_type"`
	WechatOAControls workwx.OATemplateControls `bson:"wechat_oa_controls"`
	UniqueHash       string                    `bson:"unique_hash"`
	Ctime            int64                     `bson:"ctime"`
	Utime            int64                     `bson:"utime"`
}
