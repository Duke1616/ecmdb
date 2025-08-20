package guacx

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

type Tunnel struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	Config *Config
}

func NewTunnel(address string, config *Config) (*Tunnel, error) {
	conn, err := net.DialTimeout("tcp", address, time.Second*3)
	if err != nil {
		return nil, err
	}

	tunnel := &Tunnel{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
		Config: config,
	}

	return tunnel, nil
}

func (t *Tunnel) Handshake() error {
	var err error
	defer func() {
		if err != nil {
			_ = t.conn.Close()
		}
	}()

	selectArg := t.Config.ConnectionID
	if selectArg == "" {
		selectArg = t.Config.Protocol
	}

	// select
	if _, err = t.WriteInstruction(NewInstruction("select", selectArg)); err != nil {
		return err
	}

	// args
	args, err := t.assert("args")
	if err != nil {
		return err
	}

	parameters := make([]string, len(args.Args))
	for i, name := range args.Args {
		parameters[i] = t.Config.GetParameter(name)
	}

	// size audio video image timezone
	width := t.Config.GetParameter("width")
	height := t.Config.GetParameter("height")
	dpi := t.Config.GetParameter("dpi")

	if _, err = t.WriteInstruction(NewInstruction("size", width, height, dpi)); err != nil {
		return err
	}
	if _, err = t.WriteInstruction(NewInstruction("audio", "audio/L8")); err != nil {
		return err
	}
	if _, err = t.WriteInstruction(NewInstruction("video")); err != nil {
		return err
	}
	if _, err = t.WriteInstruction(NewInstruction("image", "image/jpeg",
		"image/png", "image/webp")); err != nil {
		return err
	}
	if _, err = t.WriteInstruction(NewInstruction("timezone", "Asia/Shanghai")); err != nil {
		return err
	}

	// connect
	if _, err = t.WriteInstruction(NewInstruction("connect", parameters...)); err != nil {
		return err
	}

	// ready
	ready, err := t.assert("ready")
	if err != nil {
		return err
	}

	if len(ready.Args) == 0 {
		return fmt.Errorf("empty connection id")
	}

	t.Config.ConnectionID = ready.Args[0]
	return nil
}

func (t *Tunnel) assert(opcode string) (*Instruction, error) {
	instruction, err := t.ReadInstruction()
	if err != nil {
		return nil, err
	}

	if opcode != instruction.Opcode {
		return nil, fmt.Errorf(`expect instruction "%s" but got "%s"`, opcode, instruction.Opcode)
	}

	return instruction, nil
}

func (t *Tunnel) ReadInstruction() (*Instruction, error) {
	data, err := t.reader.ReadBytes(delimiter)
	if err != nil {
		return nil, err
	}

	instruction := (&Instruction{}).Parse(string(data))

	return instruction, nil
}

func (t *Tunnel) Write(p []byte) (int, error) {
	if t == nil || t.writer == nil {
		return 0, fmt.Errorf("cannot write to a closed tunnel")
	}
	n, err := t.writer.Write(p)
	if err != nil {
		return 0, err
	}
	err = t.writer.Flush()
	return n, err
}

func (t *Tunnel) WriteInstruction(instruction *Instruction) (int, error) {
	return t.Write(instruction.Bytes())
}

func (t *Tunnel) Close() error {
	return t.conn.Close()
}

func (t *Tunnel) Read() (p []byte, err error) {
	data, err := t.reader.ReadBytes(delimiter)
	if err != nil {
		return
	}
	s := string(data)
	if s == "rate=44100,channels=2;" {
		return make([]byte, 0), nil
	}
	if s == "rate=22050,channels=2;" {
		return make([]byte, 0), nil
	}
	if s == "5.audio,1.1,31.audio/L16;" {
		s += "rate=44100,channels=2;"
	}
	return []byte(s), err
}
