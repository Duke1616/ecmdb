package guacx

import (
	"context"

	"github.com/gorilla/websocket"
)

type GuacamoleHandler struct {
	ws     *websocket.Conn
	tunnel *Tunnel
	ctx    context.Context
	cancel context.CancelFunc
}

func NewGuacamoleHandler(ws *websocket.Conn, tunnel *Tunnel) *GuacamoleHandler {
	ctx, cancel := context.WithCancel(context.Background())
	return &GuacamoleHandler{
		ws:     ws,
		tunnel: tunnel,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (r GuacamoleHandler) Start() {
	go func() {
		for {
			select {
			case <-r.ctx.Done():
				return
			default:
				instruction, err := r.tunnel.Read()
				if err != nil {
					return
				}
				if len(instruction) == 0 {
					continue
				}
				err = r.ws.WriteMessage(websocket.TextMessage, instruction)
				if err != nil {
					return
				}
			}
		}
	}()
}

func (r GuacamoleHandler) Stop() {
	r.cancel()
	_ = r.tunnel.Close()
}
