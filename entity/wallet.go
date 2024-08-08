package entity

import (
	"context"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgp-common/lib"
)

type WalletTransaction struct {
	Transactions []runtime.WalletLedgerItem `json:"transactions"`
	Cusor        string                     `json:"cusor"`
}

func ParseWallet(payload string) (lib.Wallet, error) {
	w := lib.Wallet{}
	err := json.Unmarshal([]byte(payload), &w)
	return w, err
}

func ReadWalletUser(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger, userId string) (lib.Wallet, error) {
	return lib.ReadWalletUser(ctx, nk, logger, userId)

}

func ReadWalletUsers(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger, userIds ...string) ([]lib.Wallet, error) {
	return lib.ReadWalletUsers(ctx, nk, logger, userIds...)
}

func AddChipWalletUser(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger, userID string, wallet lib.Wallet, metadata map[string]interface{}) error {
	changeset := map[string]int64{}
	if wallet.Chips != 0 {
		changeset["chips"] = wallet.Chips // Add amountChip coins to the user's wallet.
	}
	if wallet.ChipsInBank != 0 {
		changeset["chipsInBank"] = wallet.ChipsInBank // Add amountChip coins to the user's wallet.
	}
	if wallet.Chips == 0 && wallet.ChipsInBank == 0 {
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
