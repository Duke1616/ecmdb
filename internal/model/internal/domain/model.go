package domain

import "time"

type Model struct {
	ID      int64
	GroupId int64
	Name    string
	UID     string
	Icon    string
	Builtin bool
	Ctime   time.Time
	Utime   time.Time
}

type ModelGroup struct {
	ID   int64
	Name string
}
