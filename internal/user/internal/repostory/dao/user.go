package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

const UserCollection = "c_user"

type UserDAO interface {
	CreatUser(ctx context.Context, user User) (int64, error)
	FindByUsername(ctx context.Context, username string) (User, error)
}

type userDao struct {
	db *mongox.Mongo
}

func NewUserDao(db *mongox.Mongo) UserDAO {
	return &userDao{
		db: db,
	}
}

func (dao *userDao) CreatUser(ctx context.Context, user User) (int64, error) {
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
	user.Ctime, user.Utime = now.UnixMilli(), now.UnixMilli()
	user.ID = dao.db.GetIdGenerator(UserCollection)
	col := dao.db.Collection(UserCollection)

	_, err = col.InsertOne(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	// 提交事务
	err = session.CommitTransaction(ctx)
	if err != nil {
		return 0, fmt.Errorf("提交事务错误: %w", err)
	}

	return user.ID, nil
}

func (dao *userDao) FindByUsername(ctx context.Context, username string) (User, error) {
	col := dao.db.Collection(UserCollection)
	var u User
	filter := bson.M{"username": username}

	if err := col.FindOne(ctx, filter).Decode(&u); err != nil {
		return User{}, fmt.Errorf("解码错误，%w", err)
	}

	return u, nil
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
