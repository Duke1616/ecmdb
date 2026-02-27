package scheduler

import (
	"sync"
)

type Scheduler interface {
	// Add 尝试添加任务的调度锁，若已存在则返回 false，以此避免重复向第三方或者本地 Cron 下发任务
	Add(taskId int64) bool
	// Remove 释放调度锁
	Remove(taskId int64)
}

type localScheduler struct {
	tasks sync.Map
}

func NewScheduler() Scheduler {
	return &localScheduler{
		tasks: sync.Map{},
	}
}

func (s *localScheduler) Add(taskId int64) bool {
	_, loaded := s.tasks.LoadOrStore(taskId, struct{}{})
	return !loaded // 如果是新存入的 (未被 loaded)，返回 true 表示 Add 成功获得锁
}

func (s *localScheduler) Remove(taskId int64) {
	s.tasks.Delete(taskId)
}
