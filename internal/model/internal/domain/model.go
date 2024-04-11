package domain

import "time"

type Model struct {
	ID         int64
	GroupId    int64
	Name       string
	Identifies string
	Ctime      time.Time
	Utime      time.Time
}

type ModelGroup struct {
	ID   int64
	Name string
}
