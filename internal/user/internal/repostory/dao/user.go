package dao

import "go.mongodb.org/mongo-driver/mongo"

type UserDAO interface {
}

type userDao struct {
	db *mongo.Client
}

func NewUserDao(db *mongo.Client) UserDAO {
	return userDao{
		db: db,
	}
}

type User struct {
	ID       int64  `bson:"id"`
	User     string `bson:"user"`
	Password string `bson:"password"`
	Email    string `bson:"email"`
}
