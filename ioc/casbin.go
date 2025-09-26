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
		fmt.Printf("警告: 初始化 Casbin Adapter 失败: %v\n", err)
		return nil
	}

	m, err := model.NewModelFromString(rbacModel)
	if err != nil {
		fmt.Printf("警告: Casbin 模型解析失败: %v\n", err)
		return nil
	}

	type RedisConfig struct {
		Addr     string `mapstructure:"addr"`
		DB       int    `mapstructure:"db"`
		UserName string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	}
	var cfg RedisConfig
	if err = viper.UnmarshalKey("casbin.redis", &cfg); err != nil {
		fmt.Printf("警告: 无法读取 Redis 配置: %v\n", err)
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
		fmt.Printf("警告: Enforcer 初始化失败: %v\n", err)
		return nil
	}

	_ = enforcer.SetWatcher(w)
	_ = w.SetUpdateCallback(updateCallback)

	enforcer.EnableLog(false)
	if err = enforcer.LoadPolicy(); err != nil {
		panic(err)
	}

	enforcer.StartAutoLoadPolicy(time.Minute)
	return enforcer
}

func updateCallback(rev string) {
	// 可选打印 rev
}
