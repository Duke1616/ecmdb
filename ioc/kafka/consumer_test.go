package kafka

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	consumer, err := sarama.NewConsumerGroup(addr,
		"group", cfg)
	require.NoError(t, err)

	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(time.Minute*10, func() {
		cancel()
	})
	err = consumer.Consume(ctx,
		[]string{"test_event"}, testConsumerGroupHandler{})
	t.Log(err, time.Since(start).String())
}

type testConsumerGroupHandler struct {
	ID string
}

func (t testConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	// topic => 偏移量
	//partitions := session.Claims()["idl_test_event"]
	//
	//for _, part := range partitions {
	//	session.ResetOffset("idl_test_event", part,
	//		sarama.OffsetNewest, "")
	//}
	return nil
}

func (t testConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("CleanUp")
	return nil
}

func (t testConsumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()

	const batchSize = 10
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var eg errgroup.Group
		var last *sarama.ConsumerMessage
		done := false
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					return nil
				}
				last = msg
				eg.Go(func() error {
					log.Printf("消费消息: %s\n", string(msg.Value))
					return nil
				})
			}
		}
		cancel()
		err := eg.Wait()
		if err != nil {
			continue
		}
		if last != nil {
			session.MarkMessage(last, "")
		}
	}
}

//func TestConsumerParallel(t *testing.T) {
//	cfg := sarama.NewConfig()
//
//	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
//
//	// 创建一个消费者组实例
//	consumerGroup, err := sarama.NewConsumerGroup(addr, "group", cfg)
//	require.NoError(t, err)
//
//	start := time.Now()
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	// 定时取消
//	time.AfterFunc(time.Minute*10, func() {
//		cancel()
//	})
//
//	// 启动两个消费者并发消费
//	go func() {
//		err := consumerGroup.Consume(ctx, []string{"test_event"}, &testConsumerGroupHandler{ID: "Consumer1"})
//		if err != nil {
//			log.Printf("Consumer1 error: %v\n", err)
//		}
//	}()
//
//	go func() {
//		err := consumerGroup.Consume(ctx, []string{"test_event"}, &testConsumerGroupHandler{ID: "Consumer2"})
//		if err != nil {
//			log.Printf("Consumer2 error: %v\n", err)
//		}
//	}()
//	// 等待消费者运行
//	<-ctx.Done()
//	t.Log("Consumers stopped", time.Since(start).String())
//}
