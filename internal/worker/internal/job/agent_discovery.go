package job

import (
	"context"
	"sync"

	"github.com/Duke1616/ecmdb/internal/worker/internal/service"
	"github.com/Duke1616/ework-runner/pkg/grpc/registry"
	"github.com/gotomicro/ego/core/elog"
)

const SUBSCRIBE = "agent"

type ServiceDiscoveryJob struct {
	svc      service.Service
	registry registry.Registry
	logger   *elog.Component
	close    chan struct{}

	mu sync.Mutex
	// 维护资源状态：addr -> topic, topic -> count
	agents map[string]string
	topics map[string]int
}

func NewServiceDiscoveryJob(r registry.Registry, svc service.Service) (*ServiceDiscoveryJob, error) {
	return &ServiceDiscoveryJob{
		svc:      svc,
		logger:   elog.DefaultLogger,
		registry: r,
		close:    make(chan struct{}, 1),
		agents:   make(map[string]string),
		topics:   make(map[string]int),
	}, nil
}

func (t *ServiceDiscoveryJob) Watch() {
	ctx := context.Background()

	// 1. 初始化全量数据：从 etcd 获取当前快照并同步状态
	if instances, err := t.registry.ListServices(ctx, SUBSCRIBE); err == nil {
		for _, ins := range instances {
			t.handle(ctx, registry.Event{Type: registry.EventTypeAdd, Instance: ins})
		}
	}

	// 2. 持续订阅变更：处理动态增减事件
	events := t.registry.Subscribe(SUBSCRIBE)
	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			t.handle(ctx, event)
		case <-t.close:
			return
		}
	}
}

func (t *ServiceDiscoveryJob) Close() {
	t.close <- struct{}{}
}

func (t *ServiceDiscoveryJob) handle(ctx context.Context, ev registry.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()

	addr := ev.Instance.Address
	topic, _ := ev.Instance.Metadata["topic"].(string)

	switch ev.Type {
	case registry.EventTypeAdd:
		if topic == "" || addr == "" {
			return
		}
		// 处理 Agent 的 Topic 变更或重复添加的情况
		if old, exists := t.agents[addr]; exists {
			if old == topic {
				return
			}
			t.decr(ctx, old)
		}
		t.agents[addr] = topic
		t.incr(ctx, topic)
	case registry.EventTypeDelete:
		// Delete 事件中 Metadata 可能为空，依赖内存中维护的映射关系进行定向释放
		if topic, exists := t.agents[addr]; exists {
			t.decr(ctx, topic)
			delete(t.agents, addr)
		}
	default:

	}
}

func (t *ServiceDiscoveryJob) incr(ctx context.Context, topic string) {
	if t.topics[topic]++; t.topics[topic] == 1 {
		if err := t.svc.EnsureInfrastructures(ctx, topic); err != nil {
			elog.Error("初始化基础设施失败", elog.String("topic", topic), elog.FieldErr(err))
		}
	}
}

func (t *ServiceDiscoveryJob) decr(ctx context.Context, topic string) {
	if t.topics[topic]--; t.topics[topic] == 0 {
		if err := t.svc.Release(ctx, topic); err != nil {
			elog.Error("释放基础设施失败", elog.String("topic", topic), elog.FieldErr(err))
		}
	}
}
