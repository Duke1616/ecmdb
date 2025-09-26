package ioc

import (
	"fmt"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormAdapter "github.com/casbin/gorm-adapter/v3"
	redisWatcher "github.com/casbin/redis-watcher/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

const (
	rbacModel = `[request_definition]
r = sub, obj, act, res

[policy_definition]
p = sub, obj, act, res, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act && r.res == p.res || r.sub == "root"`
)

func InitCasbin(db *gorm.DB) *casbin.SyncedEnforcer {
	adapter, err := gormAdapter.NewAdapterByDB(db)
	if err != nil {
		panic(err)
	}

	m, _ := model.NewModelFromString(rbacModel)
	type RedisConfig struct {
		Addr     string `mapstructure:"addr"`
		DB       int    `mapstructure:"db"`
		UserName string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	}
	var cfg RedisConfig
	if err = viper.UnmarshalKey("casbin.redis", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
	}

	w, err := redisWatcher.NewWatcher(cfg.Addr, redisWatcher.WatcherOptions{
		Options: redis.Options{
			DB:       cfg.DB,
			Password: cfg.Password,
		},
		Channel: "/casbin",
	})
	if err != nil {
		panic(err)
	}

	enforcer, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		panic(err)
	}

	_ = enforcer.SetWatcher(w)
	enforcer.EnableLog(false)
	if err = enforcer.LoadPolicy(); err != nil {
		panic(err)
	}

	_ = w.SetUpdateCallback(updateCallback)
	enforcer.StartAutoLoadPolicy(time.Minute)
	return enforcer
}

func updateCallback(rev string) {
	//fmt.Printf(rev, "Casbin Watcher")
}
