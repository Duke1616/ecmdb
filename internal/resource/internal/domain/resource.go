package domain

const (
	ResourceStatusRunning = iota + 1 // 运行中
	ResourceStatusStopped            // 停止
	ResourceStatusLock               // 锁定
)

const (
	ResourceTypeHosts = iota + 1
)

type Resource struct {
	ID     int64
	Type   int64
	Status int64
	Region string
	Ctime  int64
	Utime  int64
}
