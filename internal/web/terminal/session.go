package web

import (
	"sync"

	"github.com/Duke1616/ecmdb/pkg/term/guacx"
	"github.com/gorilla/websocket"
)

type Session struct {
	Websocket *websocket.Conn
	Tunnel    *guacx.Tunnel
	mutex     sync.Mutex
}

func (s *Session) Close() {
	if s.Tunnel != nil {
		_ = s.Tunnel.Close()
	}

	if s.Websocket != nil {
		_ = s.Websocket.Close()
	}
}
