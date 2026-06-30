package term

import (
	"fmt"
	"sync"
)

// SessionPool 提供并发安全的终端会话缓存池。
type SessionPool struct {
	sessions map[int64]Session
	mu       sync.RWMutex
}

// NewSessionPool 实例化一个会话池。
func NewSessionPool() *SessionPool {
	return &SessionPool{
		sessions: make(map[int64]Session),
	}
}

// GetSession 从池中读取会话（并发读安全）。
func (p *SessionPool) GetSession(id int64) (Session, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	session, exists := p.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session %d not found", id)
	}

	return session, nil
}

// SetSession 向池中存放或更新会话（写安全）。
func (p *SessionPool) SetSession(id int64, session Session) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.sessions[id] = session
}

// DeleteSession 从池中注销并清除某个会话。
func (p *SessionPool) DeleteSession(id int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.sessions, id)
}
