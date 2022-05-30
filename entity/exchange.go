package entity

import "time"

type ExchangeListCursor struct {
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
