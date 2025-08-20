package sshx

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type SSHConnect struct {
	tick         *time.Ticker
	conn         *websocket.Conn
	ctx          context.Context
	cancel       context.CancelFunc
	session      *ssh.Session
	StdinPipe    io.WriteCloser
	stdoutReader *bufio.Reader
	dataChan     chan rune
	mutex        sync.Mutex
	buf          bytes.Buffer
}

func (s *SSHConnect) send() {
	defer func() {
		s.buf.Reset()
	}()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.tick.C:
			msg := s.buf.String()
			if msg != "" {
				if err := s.SendMessageToWebSocket(msg); err != nil {
					return
				}
				s.buf.Reset()
			}
		case data := <-s.dataChan:
			p := make([]byte, utf8.RuneLen(data))
			utf8.EncodeRune(p, data)
			s.buf.Write(p)
		}
	}
}

func (s *SSHConnect) output() {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			rn, size, err := s.stdoutReader.ReadRune()
			if err != nil {
				return
			}

			if size == 0 {
				continue
			}

			if rn == utf8.RuneError {
				continue
			}

			s.dataChan <- rn
		}
	}
}

func (s *SSHConnect) Start() {
	go s.send()
	go s.output()
}

func (s *SSHConnect) SendMessageToWebSocket(msg string) error {
	message, err := json.Marshal(TerminalMessage{Operation: "stdout", Data: msg})
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.conn.WriteMessage(websocket.TextMessage, message)
}

func NewSSHConnect(client *ssh.Client, conn *websocket.Conn, height, width int) (sshConn *SSHConnect, err error) {
	var (
		session *ssh.Session
	)
	if session, err = client.NewSession(); err != nil {
		return
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err = session.RequestPty("xterm-256color", height, width, modes); err != nil {
		return nil, err
	}

	var pipe io.WriteCloser
	if pipe, err = session.StdinPipe(); err != nil {
		return nil, err
	}

	var stdoutPipe io.Reader
	if stdoutPipe, err = session.StdoutPipe(); err != nil {
		return nil, err
	}
	stdoutReader := bufio.NewReader(stdoutPipe)

	if err = session.Shell(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &SSHConnect{
		session:      session,
		tick:         time.NewTicker(60 * time.Millisecond),
		ctx:          ctx,
		conn:         conn,
		cancel:       cancel,
		dataChan:     make(chan rune),
		StdinPipe:    pipe,
		stdoutReader: stdoutReader,
	}, nil
}

func (s *SSHConnect) WindowChange(h int, w int) error {
	return s.session.WindowChange(h, w)
}

func (s *SSHConnect) Stop() {
	s.cancel()
	s.tick.Stop()
	close(s.dataChan)

	// 关闭会话
	if s.session != nil {
		_ = s.session.Close()
	}
}
