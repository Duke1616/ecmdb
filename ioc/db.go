package ioc

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/Duke1616/ecmdb/pkg/mongox/plugin"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongoDB() *mongox.Mongo {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			//fmt.Println(evt.Command)
		},
	}

	type Config struct {
		DSN      string `mapstructure:"dsn"`
		DB       string `mapstructure:"db"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("mongodb", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into structure: %v", err))
	}
	dsn := strings.Split(cfg.DSN, "//")
	uri := fmt.Sprintf("%s//%s:%s@%s", dsn[0], cfg.Username, cfg.Password, dsn[1])

	opts := options.Client().
		ApplyURI(uri).
		SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)

	if err != nil {
		panic(err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Panicf("ping mongodb server error, %s", err)
	}

	return mongox.NewMongo(client, cfg.DB)
}

func InitMongoDBV2(db *mongox.Mongo) *mongox.DB {
	client := db.DBClient
	dbName := db.Database().Name()
	dbV2 := mongox.NewDB(client, dbName)

	dbV2.Use(plugin.NewAutoIDPlugin(dbV2.Database()))
	dbV2.Use(plugin.NewTenantPlugin())

	return dbV2
}
