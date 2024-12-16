package web

import (
	"github.com/Duke1616/ecmdb/pkg/guacx"
	"github.com/gorilla/websocket"
	"sync"
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
