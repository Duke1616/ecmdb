//go:build wireinject

package rota

import (
	"fmt"
	"github.com/Duke1616/ecmdb/internal/rota/internal/repository"
	"github.com/Duke1616/ecmdb/internal/rota/internal/repository/dao"
	"github.com/Duke1616/ecmdb/internal/rota/internal/service"
	"github.com/Duke1616/ecmdb/internal/rota/internal/service/schedule"
	"github.com/Duke1616/ecmdb/internal/rota/internal/web"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/google/wire"
	"time"
)

var ProviderSet = wire.NewSet(
	web.NewHandler,
	service.NewService,
	repository.NewRotaRepository,
	dao.NewRotaDao,
)

func InitModule(db *mongox.Mongo) (*Module, error) {
	wire.Build(
		ProviderSet,
		InitScheduleRule,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

func InitScheduleRule() schedule.Scheduler {
	// 创建一个位置对象，表示中国北京的位置
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		fmt.Print()
	}

	return schedule.NewRruleSchedule(location)
}
