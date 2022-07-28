package entity

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
)

type Wallet struct {
	UserId      string
	Chips       int64 `json:"chips"`
	ChipsInBank int64 `json:"chipsInbank"`
}

type WalletTransaction struct {
	Transactions []runtime.WalletLedgerItem `json:"transactions"`
	Cusor        string                     `json:"cusor"`
}

func ParseWallet(payload string) (Wallet, error) {
	w := Wallet{}
	err := json.Unmarshal([]byte(payload), &w)
	return w, err
}

func ReadWalletUser(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger, userId string) (Wallet, error) {
	// logger.Info("Read wauserIds %v", nk, ctx, userIds)
	account, err := nk.AccountGetId(ctx, userId)
	if err != nil {
		logger.Error("Error when read account, error: %s, userId %s",
			err.Error(), userId)
		return Wallet{}, err
	}
	w, e := ParseWallet(account.Wallet)
	return w, e
}

func ReadWalletUsers(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger, userIds ...string) ([]Wallet, error) {
	// logger.Info("Read wauserIds %v", nk, ctx, userIds)
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

func AddChipWalletUser(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger, userID string, wallet Wallet, metadata map[string]interface{}) error {
	changeset := map[string]int64{}
	if wallet.Chips != 0 {
		changeset["chips"] = wallet.Chips // Add amountChip coins to the user's wallet.
	}
	if wallet.ChipsInBank != 0 {
		changeset["chipsInBank"] = wallet.ChipsInBank // Add amountChip coins to the user's wallet.
	}
	if wallet.Chips <= 0 && wallet.ChipsInBank <= 0 {
		return nil
	}
	// metadata := map[string]interface{}{
	// 	"game_topup": reason,
	// }

	_, _, err := nk.WalletUpdate(ctx, userID, changeset, metadata, true)
	if err != nil {
		logger.WithField("err", err).Error("Wallet update error.")
	}
	return err
}
