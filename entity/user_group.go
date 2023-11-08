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

type UserGroupUserInfo struct {
	Level       int64
	VipLevel    int64
	AG          int64
	ChipsInBank int64
	Co          int64
	CO0         int64
	LQ          int64
	BLQ1        int64
	BLQ3        int64
	BLQ5        int64
	BLQ7        int64
	Avgtrans7   int64
	CreateTime  int64
}
