package entity

import (
	"github.com/heroiclabs/nakama-common/api"
)

type Account struct {
	api.Account
	LastOnlineTimeUnix int64
	Sid                int64
}
