package strategy

import (
	"context"
	"fmt"
	"sync"

	"github.com/Duke1616/enotify/notify"
)

const (
	// ProcessEndSend 流程结束后发送
	ProcessEndSend = 1
	// ProcessNowSend 当前节点通过直接发送
	ProcessNowSend = 2
)

func send(ctx context.Context, notifyWrap []notify.NotifierWrap) (bool, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstError error
	success := true

	// 使用 goroutines 发送消息
	for _, msg := range notifyWrap {
		wg.Add(1)
		nw := msg
		go func(m notify.NotifierWrap) {
			defer wg.Done()

			ok, err := nw.Send(ctx)
			if err != nil {
				mu.Lock() // 锁定访问共享资源
				if firstError == nil {
					firstError = err // 记录第一个出现的错误
				}
				success = false
				mu.Unlock()
			}

			if !ok {
				mu.Lock()
				if firstError == nil {
					firstError = fmt.Errorf("消息发送失败")
				}
				success = false
				mu.Unlock()
			}
		}(msg)
	}

	// 等待所有 goroutine 完成
	wg.Wait()

	if firstError != nil {
		return false, firstError
	}
	return success, nil
}
