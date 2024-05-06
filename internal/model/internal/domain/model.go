package domain

import "time"

type Model struct {
	ID      int64
	GroupId int64
	Name    string
	UID     string
	Icon    string
	Ctime   time.Time
	Utime   time.Time
}

type ModelGroup struct {
	ID   int64
	Name string
}

type ModelsByGroupId struct {
	ID     int64
	Name   string
	Models []Model
}
