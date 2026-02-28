package dao

import (
	"context"
	"time"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type FavoriteDAO interface {
	// ToggleFavorite 切换收藏状态，返回最新状态 (true 为收藏，false 为取消)
	ToggleFavorite(ctx context.Context, userId int64, templateId int64) (bool, error)
	// ListTemplateIdsByUserId 获取指定用户收藏的模版 ID 列表
	ListTemplateIdsByUserId(ctx context.Context, userId int64) ([]int64, error)
}

type favoriteDAO struct {
	db *mongox.Mongo
}

func NewFavoriteDAO(db *mongox.Mongo) FavoriteDAO {
	return &favoriteDAO{
		db: db,
	}
}

func (dao *favoriteDAO) ToggleFavorite(ctx context.Context, userId int64, templateId int64) (bool, error) {
	col := dao.db.Collection(TemplateFavoriteCollection)
	now := time.Now().UnixMilli()
	doc := TemplateFavorite{
		Id:         dao.db.GetIdGenerator(TemplateFavoriteCollection),
		UserId:     userId,
		TemplateId: templateId,
		Ctime:      now,
		Utime:      now,
	}

	_, err := col.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			filter := bson.M{"user_id": userId, "template_id": templateId}
			_, delErr := col.DeleteOne(ctx, filter)
			return false, delErr
		}
		return false, err
	}

	return true, nil
}

func (dao *favoriteDAO) ListTemplateIdsByUserId(ctx context.Context, userId int64) ([]int64, error) {
	col := dao.db.Collection(TemplateFavoriteCollection)
	filter := bson.M{"user_id": userId}

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var favs []TemplateFavorite
	if err = cursor.All(ctx, &favs); err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(favs))
	for _, f := range favs {
		ids = append(ids, f.TemplateId)
	}

	return ids, nil
}
