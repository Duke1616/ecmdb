package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"time"
)

const MenuCollection = "c_menu"

type MenuDAO interface {
	CreateMenu(ctx context.Context, t Menu) (int64, error)
}

type menuDAO struct {
	db *mongox.Mongo
}

func (dao *menuDAO) CreateMenu(ctx context.Context, e Menu) (int64, error) {
	e.Id = dao.db.GetIdGenerator(MenuCollection)
	col := dao.db.Collection(MenuCollection)
	now := time.Now()
	e.Ctime, e.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, e)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return e.Id, nil
}

func NewMenuDAO(db *mongox.Mongo) MenuDAO {
	return &menuDAO{
		db: db,
	}
}

type Menu struct {
	Id          int64   `bson:"id"`
	Pid         int64   `bson:"pid"`
	Name        string  `bson:"name"`
	Path        string  `bson:"path"`
	Sort        int64   `bson:"sort"`
	IsRoot      bool    `bson:"is_root"`
	Type        uint8   `bson:"type"`
	Meta        Meta    `bson:"meta"`
	EndpointIds []int64 `bson:"endpoint_ids"`
	Ctime       int64   `bson:"ctime"`
	Utime       int64   `bson:"utime"`
}

type Meta struct {
	Title       string `bson:"title"`        // 展示名称
	IsHidden    bool   `bson:"is_hidden"`    // 是否展示
	IsAffix     bool   `bson:"is_affix"`     // 是否固定
	IsKeepAlive bool   `bson:"is_keepalive"` // 是否缓存
	Icon        string `bson:"icon"`         // Icon图标
}
