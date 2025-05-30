package entity

import (
	"time"
)

type FeeGame struct {
	Id             int64
	UserID         string
	Game           string
	Fee            int64
	CreateTimeUnix int64
	From           int64
	To             int64
}

type FeeGameListCursor struct {
	Id         string
	Offset     int64
	Limit      int64
	UserId     string
	CreateTime time.Time
	IsNext     bool
	Total      int64
	From       int64
	To         int64
}
