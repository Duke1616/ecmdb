package ioc

import (
	"context"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func InitMongoDB() *mongox.Mongo {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			//fmt.Println(evt.Command)
		},
	}

	opts := options.Client().
		ApplyURI("mongodb://cmdb:123456@10.31.0.200:47017/cmdb").
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
