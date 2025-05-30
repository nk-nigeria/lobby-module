package entity

import (
	"time"
)

type NotificationListCursor struct {
	Id         int64
	UserId     string
	Offset     int64
	CreateTime time.Time
	IsNext     bool
	Total      int64
}
