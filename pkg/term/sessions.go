package term

import (
	"fmt"
	"sync"

	"golang.org/x/crypto/ssh"
)

type Sessions struct {
	SshClient *ssh.Client
}

func NewSessions(sshClient *ssh.Client) *Sessions {
	return &Sessions{
		SshClient: sshClient,
	}
}

type SessionPool struct {
	sessions map[int64]*Sessions
	mu       *sync.Mutex
}

func NewSessionPool() *SessionPool {
	return &SessionPool{
		sessions: make(map[int64]*Sessions),
		mu:       &sync.Mutex{},
	}
}

func (p *SessionPool) GetSession(id int64) (*Sessions, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	session, exists := p.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session %d already exists", id)
	}

	return session, nil
}

func (p *SessionPool) SetSession(id int64, session *Sessions) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.sessions[id] = session
}
