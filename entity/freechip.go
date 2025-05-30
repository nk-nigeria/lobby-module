package entity

import (
	"time"
)

type FreeChipListCursor struct {
	Id          int64
	Offset      int64
	UserId      string
	CreateTime  time.Time
	IsNext      bool
	Total       int64
	ClaimStatus int
}
