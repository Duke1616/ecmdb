package sshx

import "encoding/json"

type TerminalMessage struct {
	Operation string `json:"operation"`
	Data      string `json:"data"`
	Cols      int    `json:"cols"`
	Rows      int    `json:"rows"`
}

func ParseTerminalMessage(value []byte) (TerminalMessage, error) {
	m := TerminalMessage{}
	err := json.Unmarshal(value, &m)
	return NewMessage(m.Operation, m.Data, m.Cols, m.Rows), err
}

func NewMessage(operation string, data string, cols, rows int) TerminalMessage {
	return TerminalMessage{Operation: operation, Data: data, Cols: cols, Rows: rows}
}
