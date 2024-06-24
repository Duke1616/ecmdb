package ioc

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strings"
	"time"
)

func InitMongoDB(viper *viper.Viper) *mongox.Mongo {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			fmt.Println(evt.Command)
		},
	}

	type Config struct {
		DSN      string `mapstructure:"dsn"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("mongodb", &cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
	}

	dns := strings.Split(cfg.DSN, "//")
	uri := fmt.Sprintf("%s//%s:%s@%s", dns[0], cfg.Username, cfg.Password, dns[1])
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

	return mongox.NewMongo(client, "cmdb")
}
