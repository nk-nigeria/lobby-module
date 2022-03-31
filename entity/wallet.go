package entity

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
)

type Wallet struct {
	UserId string
	Chips  int64 `json:"chips"`
}

func ParseWallet(payload string) (Wallet, error) {
	w := Wallet{}
	err := json.Unmarshal([]byte(payload), &w)
	return w, err
}

func ReadWalletUsers(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger, userIds ...string) ([]Wallet, error) {
	logger.Error("nk %v ctx %v userIds %v", nk, ctx, userIds)
	accounts, err := nk.AccountsGetId(ctx, userIds)
	if err != nil {
		logger.Error("Error when read list account, error: %s, list userId %s",
			err.Error(), strings.Join(userIds, ","))
		return nil, err
	}
	wallets := make([]Wallet, 0)
	for _, ac := range accounts {
		w, e := ParseWallet(ac.Wallet)
		if e != nil {
			logger.Error("Error when parse wallet user %s, error: %s", ac.User.Id, e.Error())
			return wallets, err
		}
		w.UserId = ac.User.Id
		wallets = append(wallets, w)
	}
	return wallets, nil
}
