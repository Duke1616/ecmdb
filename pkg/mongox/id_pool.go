package mongox

import (
	"context"
	"fmt"
	"sync"
)

// IDPool ID池，用于高并发场景的性能优化
type IDPool struct {
	mongo          *Mongo
	collectionName string
	poolSize       int
	minPoolSize    int // 最小池大小
	maxPoolSize    int // 最大池大小
	mu             sync.Mutex
	currentID      int64
	maxID          int64
	refillCount    int64 // 重新填充次数（用于自适应调整）
}

// NewIDPool 创建一个ID池
// poolSize: 每次从数据库获取多少个ID（例如1000）
func NewIDPool(mongo *Mongo, collectionName string, poolSize int) *IDPool {
	if poolSize <= 0 {
		poolSize = 100 // 默认预分配100个ID（较小的值减少浪费）
	}
	return &IDPool{
		mongo:          mongo,
		collectionName: collectionName,
		poolSize:       poolSize,
		minPoolSize:    100,   // 最小100个
		maxPoolSize:    10000, // 最大10000个
		currentID:      0,
		maxID:          0,
		refillCount:    0,
	}
}

// NewIDPoolWithRange 创建带自定义范围的ID池
func NewIDPoolWithRange(mongo *Mongo, collectionName string, minSize, maxSize int) *IDPool {
	if minSize <= 0 {
		minSize = 100
	}
	if maxSize < minSize {
		maxSize = minSize
	}
	return &IDPool{
		mongo:          mongo,
		collectionName: collectionName,
		poolSize:       minSize,
		minPoolSize:    minSize,
		maxPoolSize:    maxSize,
		currentID:      0,
		maxID:          0,
		refillCount:    0,
	}
}

// NextID 获取下一个ID（线程安全）
func (p *IDPool) NextID(ctx context.Context) (int64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 如果池中的ID用完了，重新批量获取
	if p.currentID >= p.maxID {
		// 自适应调整池大小
		p.adaptPoolSize()

		startID, err := p.mongo.GetBatchIdGenerator(p.collectionName, p.poolSize)
		if err != nil {
			return 0, fmt.Errorf("重新填充ID池失败: %w", err)
		}
		p.currentID = startID
		p.maxID = startID + int64(p.poolSize)
		p.refillCount++
	}

	// 从池中取一个ID
	id := p.currentID
	p.currentID++
	return id, nil
}

// adaptPoolSize 自适应调整池大小（需要在持锁状态下调用）
// 如果频繁重新填充，说明使用量大，增加池大小
// 如果很少重新填充，说明使用量小，减少池大小（减少浪费）
func (p *IDPool) adaptPoolSize() {
	// 每10次重新填充，评估一次池大小
	if p.refillCount > 0 && p.refillCount%10 == 0 {
		// 简单策略：如果频繁重填，增大池；否则减小
		// 实际可以根据时间间隔、使用率等更复杂的指标调整
		if p.poolSize < p.maxPoolSize {
			p.poolSize = min(p.poolSize*2, p.maxPoolSize)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NextBatch 批量获取ID（线程安全）
func (p *IDPool) NextBatch(ctx context.Context, count int) (startID int64, err error) {
	if count <= 0 {
		return 0, fmt.Errorf("count must be positive")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 如果请求的数量超过池大小，直接从数据库获取
	if count > p.poolSize {
		return p.mongo.GetBatchIdGenerator(p.collectionName, count)
	}

	// 如果池中剩余的ID不够，重新填充
	if p.currentID+int64(count) > p.maxID {
		newStartID, err := p.mongo.GetBatchIdGenerator(p.collectionName, p.poolSize)
		if err != nil {
			return 0, fmt.Errorf("重新填充ID池失败: %w", err)
		}
		p.currentID = newStartID
		p.maxID = newStartID + int64(p.poolSize)
	}

	// 从池中取出一批ID
	startID = p.currentID
	p.currentID += int64(count)
	return startID, nil
}

// Stats 返回ID池的统计信息
func (p *IDPool) Stats() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	return map[string]interface{}{
		"collection_name": p.collectionName,
		"pool_size":       p.poolSize,
		"min_pool_size":   p.minPoolSize,
		"max_pool_size":   p.maxPoolSize,
		"current_id":      p.currentID,
		"max_id":          p.maxID,
		"remaining":       p.maxID - p.currentID,
		"refill_count":    p.refillCount,
		"waste_on_close":  p.maxID - p.currentID, // 如果现在关闭会浪费多少ID
	}
}

// Close 优雅关闭ID池
// 尝试将未使用的ID归还到数据库（可选，用于减少浪费）
// 注意：这个操作有风险，只在优雅关闭时调用，crash时无法执行
func (p *IDPool) Close(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 计算未使用的ID数量
	unused := p.maxID - p.currentID
	if unused <= 0 {
		return nil // 没有未使用的ID
	}

	// 尝试归还未使用的ID到数据库
	// 这里使用负数来减少 next_id
	coll := p.mongo.Database().Collection("c_id_generator")
	filter := map[string]interface{}{"name": p.collectionName}
	update := map[string]interface{}{
		"$inc": map[string]interface{}{"next_id": -unused},
	}

	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("归还未使用的ID失败: %w", err)
	}

	// 重置池状态
	p.currentID = 0
	p.maxID = 0

	return nil
}

// Remaining 返回池中剩余的ID数量
func (p *IDPool) Remaining() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.maxID - p.currentID
}
