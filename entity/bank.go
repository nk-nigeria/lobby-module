package entity

import (
	"context"
	"errors"

	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	CollectionBank = "collection_bank"
)

func BankPushToSafe(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, unmarshaler *protojson.UnmarshalOptions, bank *pb.Bank) (*pb.Bank, error) {
	if bank.GetAmoutChip() <= 0 {
		return nil, errors.New("Amout chip must larger than zero")
	}

	wallets, err := ReadWalletUsers(ctx, nk, logger, bank.GetSenderId())
	if err != nil {
		return nil, err
	}
	if len(wallets) == 0 {
		return nil, errors.New("User not have wallet")
	}
	currentWallet := wallets[0]
	if currentWallet.Chips < bank.AmoutChip {
		return nil, errors.New("User chips smaller than amout chip push to safe")
	}
	newUserWallet := currentWallet
	newUserWallet.Chips = -bank.AmoutChip
	newUserWallet.ChipsInBank = bank.AmoutChip
	// substract chip in wallet
	err = AddChipWalletUser(ctx, nk, logger, newUserWallet.UserId, newUserWallet, "push_to_safe")
	if err != nil {
		logger.Error("Update wallet when push to safe action error: %s", err.Error())
		return nil, err
	}
	// add chip in bank
	newBank := bank
	newBank.AmoutChip = newUserWallet.ChipsInBank
	logger.Info("User id %s push %d to safe, wallet before: chips %d, bank: %d, wallets after: chips %d, bank %d",
		bank.GetSenderId(), bank.AmoutChip, currentWallet.Chips,
		currentWallet.ChipsInBank, currentWallet.Chips-bank.AmoutChip,
		currentWallet.ChipsInBank+bank.AmoutChip)
	return newBank, nil
}

func BankWithdraw(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, bank *pb.Bank) (*pb.Bank, error) {
	if bank.GetAmoutChip() <= 0 {
		return nil, errors.New("Amout chip must larger than zero")
	}

	wallets, err := ReadWalletUsers(ctx, nk, logger, bank.GetSenderId())
	if err != nil {
		return nil, err
	}
	if len(wallets) == 0 {
		return nil, errors.New("User not have wallet")
	}
	currentWallet := wallets[0]
	if currentWallet.ChipsInBank < bank.AmoutChip {
		return nil, errors.New("User chips smaller than amout chip push to safe")
	}

	newUserWallet := Wallet{}
	newUserWallet.Chips = bank.AmoutChip
	newUserWallet.ChipsInBank = -bank.AmoutChip
	err = AddChipWalletUser(ctx, nk, logger, bank.SenderId, newUserWallet, "with_draw")
	if err != nil {
		logger.Error("Withdraw to wallet user %s, amout chip %d, err: %s", bank.GetSenderId(), bank.GetAmountFee(), err.Error())
		return nil, err
	}
	newBank := bank
	newBank.AmoutChip = currentWallet.ChipsInBank - bank.AmoutChip
	logger.Info("User id %s push %d to safe, wallet before: chips %d, bank: %d, wallets after: chips %d, bank %d",
		bank.GetSenderId(), currentWallet.Chips,
		currentWallet.ChipsInBank, currentWallet.Chips+bank.AmoutChip,
		currentWallet.ChipsInBank-bank.AmoutChip)
	return newBank, nil
}

func BankSendGift(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, bank *pb.Bank) (*pb.Bank, error) {
	if bank.GetAmoutChip() <= 0 {
		return nil, errors.New("Amout chip must larger than zero")
	}
	if bank.GetSenderId() == bank.GetRecipientId() {
		return nil, errors.New("Reciver must diffirent sender")
	}

	wallets, err := ReadWalletUsers(ctx, nk, logger, bank.GetSenderId())
	if err != nil {
		return nil, err
	}
	if len(wallets) == 0 {
		return nil, errors.New("User not have wallet")
	}
	senderWallet := wallets[0]

	if senderWallet.ChipsInBank < bank.AmoutChip {
		logger.Error("Sender %s amout chip smaller than amout chip request send gift, chips in bank: %d, request send gift: %d",
			bank.GetSenderId(), senderWallet.ChipsInBank, bank.AmoutChip)
		err = errors.New("Sender amout chip smaller than amout chip request send gift")
		return nil, err
	}

	senderNewWallet := Wallet{}
	senderNewWallet.ChipsInBank = -bank.AmoutChip - bank.AmountFee
	err = AddChipWalletUser(ctx, nk, logger,
		bank.GetSenderId(),
		senderNewWallet,
		"send_gift")
	if err != nil {
		logger.Error("Update wallet sender %s error: %s", bank.GetSenderId(), err.Error())
		return nil, err
	}
	reciverNewWallet := Wallet{}
	reciverNewWallet.ChipsInBank = bank.AmoutChip
	// add chip recv wallet
	err = AddChipWalletUser(ctx, nk, logger,
		bank.GetRecipientId(),
		reciverNewWallet,
		"recv_gift")
	if err != nil {
		logger.Error("Update wallet recv %s error: %s", bank.GetSenderId(), err.Error())
		// revert sender wallet
		revertSenderWallet := Wallet{}
		revertSenderWallet.ChipsInBank = -senderNewWallet.ChipsInBank
		if e := AddChipWalletUser(ctx, nk, logger,
			bank.GetSenderId(),
			revertSenderWallet,
			"send_gift_revert"); e != nil {
			logger.Error("Revert sender wallet %s error:%s", bank.GetSenderId(), e.Error())
		}
		return nil, err
	}
	logger.Info("Sender %s, send %d chips --> recv %s, fee %d (%d chips)",
		bank.GetSenderId(), bank.AmoutChip,
		bank.RecipientId, bank.PercenFee, bank.AmountFee)
	return nil, nil
}

func BankHistory(logger runtime.Logger, nk runtime.NakamaModule, userID string, limit, offset int64) {
}
