package domain

import "time"

type Model struct {
	ID      int64
	GroupId int64
	Name    string
	UID     string
	Ctime   time.Time
	Utime   time.Time
}

type ModelGroup struct {
	ID   int64
	Name string
}
