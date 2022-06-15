package entity

import (
	"time"
)

type InAppMessageListCursor struct {
	Id         int64
	Offset     int64
	CreateTime time.Time
	IsNext     bool
	Total      int64
}
