package kafka

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
)

func TestAsyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Errors = true
	cfg.Producer.Return.Successes = true
	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Idempotent = true       // 启用幂等性
	cfg.Producer.Return.Successes = true // 需要成功的消息
	cfg.Net.MaxOpenRequests = 1

	producer, err := sarama.NewAsyncProducer(addr, cfg)
	require.NoError(t, err)
	msgCh := producer.Input()
	go func() {
		var counter int64
		for {
			time.Sleep(2 * time.Second)

			atomic.AddInt64(&counter, 1)
			msg := &sarama.ProducerMessage{
				Topic: "test_event",
				Key:   sarama.StringEncoder(fmt.Sprintf("idl-%d", counter)), // 使用动态 key 来确保均匀分配
				Value: sarama.StringEncoder(fmt.Sprintf("Hello, 这是一条消息 %d", counter)),
				Headers: []sarama.RecordHeader{
					{
						Key:   []byte("trace_id"),
						Value: []byte("123456"),
					},
				},
				Metadata: "这是metadata",
			}
			select {
			case msgCh <- msg:

			}

		}
	}()

	errCh := producer.Errors()
	successCh := producer.Successes()

	for {
		// 如果两个情况都没发生，就会阻塞
		select {
		case err := <-errCh:
			t.Log("发送出了问题", err.Err)
		case <-successCh:
			t.Log("发送成功")
		}
	}
}
