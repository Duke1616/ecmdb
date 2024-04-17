package dao

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

const UserCollection = "c_user"

type UserDAO interface {
	CreatUser(ctx context.Context, user User) (int64, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
}

type userDao struct {
	db *mongox.Mongo
}

func NewUserDao(client *mongo.Client) UserDAO {
	return &userDao{
		db: mongox.NewMongo(client),
	}
}

func (dao *userDao) CreatUser(ctx context.Context, user User) (int64, error) {
	now := time.Now()
	user.Ctime, user.Utime = now.UnixMilli(), now.UnixMilli()
	user.ID = dao.db.GetIdGenerator(UserCollection)
	col := dao.db.Collection(UserCollection)

	_, err := col.InsertMany(ctx, []interface{}{user})

	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

func (dao *userDao) FindByUsername(ctx context.Context, username string) (*User, error) {
	col := dao.db.Collection(UserCollection)
	m := &User{}
	filter := bson.M{"username": username}

	if err := col.FindOne(ctx, filter).Decode(m); err != nil {
		return &User{}, err
	}

	return m, nil
}

// 账号来源
const (
	Ldap   = iota + 1 // LDAP 创建
	System            // 系统 创建
)

// 创建用户方式
const (
	LdapSync = iota + 1
	UserRegistry
)

type User struct {
	ID         int64  `bson:"id"`
	Username   string `bson:"username"`
	Password   string `bson:"password"`
	Email      string `bson:"email"`
	Title      string `json:"title"`
	SourceType int64  `json:"source_type"`
	CreateType int64  `json:"create_type"`
	Ctime      int64  `bson:"ctime"`
	Utime      int64  `bson:"utime"`
}
