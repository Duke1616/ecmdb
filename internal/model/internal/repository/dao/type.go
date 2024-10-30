package dao

const (
	ModelCollection      = "c_model"
	ModelGroupCollection = "c_model_group"
)

type Model struct {
	Id           int64  `bson:"id"`
	ModelGroupId int64  `bson:"model_group_id"`
	Name         string `bson:"name"`
	UID          string `bson:"uid"`
	Icon         string `bson:"icon"`
	Ctime        int64  `bson:"ctime"`
	Utime        int64  `bson:"utime"`
}

type ModelGroup struct {
	Id    int64  `bson:"id"`
	Name  string `bson:"name"`
	Ctime int64  `bson:"ctime"`
	Utime int64  `bson:"utime"`
}
