package entity

import (
	"time"
)

type UserGroupListCursor struct {
	Id         int64
	Offset     int64
	CreateTime time.Time
	IsNext     bool
	Total      int64
}
