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
	currentBank, err := ReadBank(ctx, nk, logger, bank.GetSenderId())
	if err != nil {
		logger.Error("Read bank user %s, err: %s", bank.GetSenderId(), err.Error())
		return nil, err
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
	newUserWallet.Chips = currentWallet.Chips - bank.AmoutChip
	// substract chip in wallet
	err = AddChipWalletUser(ctx, nk, logger, newUserWallet.UserId, -bank.GetAmoutChip(), "push_to_safe")
	if err != nil {
		logger.Error("Update wallet when push to safe action error: %s", err.Error())
		return nil, err
	}
	// add chip in bank
	newBank := currentBank
	newBank.AmoutChip = currentBank.AmoutChip + bank.AmoutChip
	err = UpdateBank(ctx, nk, logger, newBank)
	if err != nil {
		logger.Error("Push to bank from user %s, err: %s, revert wallet", bank.GetSenderId(), err.Error())
		err = AddChipWalletUser(ctx, nk, logger, newUserWallet.UserId, bank.GetAmoutChip(), "revert_push_to_safe")
		if err != nil {
			logger.Error("Revert wallet for user %s err: %s", newUserWallet.UserId, err.Error())
		}
		return nil, err
	}
	logger.Info("User id %s push %d to safe, wallet chips before %d, wallets chip after %d, bank before %d, bank after %d",
		bank.GetSenderId(), bank.GetAmoutChip(),
		currentWallet.Chips, newUserWallet.Chips,
		currentBank.AmoutChip, newBank.AmoutChip)
	return newBank, nil
}

func BankWithdraw(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, bank *pb.Bank) (*pb.Bank, error) {
	if bank.GetAmoutChip() <= 0 {
		return nil, errors.New("Amout chip must larger than zero")
	}
	currentBank, err := ReadBank(ctx, nk, logger, bank.GetSenderId())
	if err != nil {
		logger.Error("Read bank user %s, err: %s", bank.GetSenderId(), err.Error())
		return nil, err
	}
	if currentBank.AmoutChip < bank.AmoutChip {
		err = errors.New("Amout chip in bank smaller than withdraw request")
		logger.Error("User %d, request withdraw lager than amount in bank, current %d, req withdraw %d",
			bank.GetSenderId(), currentBank.AmoutChip, bank.AmoutChip)
		return nil, err
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

	err = AddChipWalletUser(ctx, nk, logger, bank.SenderId, bank.AmoutChip, "with_draw")
	if err != nil {
		logger.Error("Withdraw to wallet user %s, amout chip %d, err: %s", bank.GetSenderId(), bank.GetAmountFee(), err.Error())
		return nil, err
	}
	newUserWallet := currentWallet
	newUserWallet.Chips = currentBank.AmoutChip + bank.AmoutChip
	newBank := currentBank
	newBank.AmoutChip = newBank.AmoutChip - bank.AmoutChip
	err = UpdateBank(ctx, nk, logger, newBank)
	if err != nil {
		logger.Error("Withdraw from user %s, err: %s, revert wallet", bank.GetSenderId(), err.Error())
		err = AddChipWalletUser(ctx, nk, logger, bank.SenderId, -bank.GetAmoutChip(), "revert_with_draw")
		if err != nil {
			logger.Error("Revert wallet for user %s err: %s", bank.GetSenderId(), err.Error())
		}
		return nil, err
	}
	logger.Info("User id %s with draw %d, wallet chips before %d, wallets chip after %d, bank before %d, bank after %d",
		bank.GetSenderId(), bank.GetAmoutChip(),
		currentWallet.Chips, newUserWallet.Chips,
		currentBank.AmoutChip, newBank.AmoutChip)
	return newBank, nil
}

func BankSendGift(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, bank *pb.Bank) (*pb.Bank, error) {
	if bank.GetAmoutChip() <= 0 {
		return nil, errors.New("Amout chip must larger than zero")
	}
	if bank.GetSenderId() == bank.GetRecipientId() {
		return nil, errors.New("Reciver must diffirent sender")
	}

	senderBank, err := ReadBank(ctx, nk, logger, bank.GetSenderId())
	if err != nil {
		logger.Error("Read sender bank %s err: %s", bank.GetSenderId(), err.Error())
		return nil, err
	}
	if senderBank.AmoutChip < bank.AmoutChip {
		err = errors.New("Sender amout chip smaller than amout chip request send gift")
		return nil, err
	}
	reciverBank, err := ReadBank(ctx, nk, logger, bank.GetRecipientId())
	if err != nil {
		logger.Error("Read reciver bank %s err: %s", bank.GetRecipientId(), err.Error())
		return nil, err
	}
	senderNewBank := senderBank
	senderBank.AmoutChip = senderBank.AmoutChip - bank.AmoutChip - bank.AmountFee
	err = UpdateBank(ctx, nk, logger, senderNewBank)
	if err != nil {
		logger.Error("Update sender %s error: %s", senderBank.GetSenderId(), err.Error())
		return nil, err
	}
	reciverNewBank := reciverBank
	reciverNewBank.AmoutChip = reciverBank.AmoutChip + bank.AmoutChip
	return nil, nil
}

func BankHistory(logger runtime.Logger, nk runtime.NakamaModule, userID string, limit, offset int64) {
}

func UpdateBank(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger, bank *pb.Bank) error {
	marshaler := &protojson.MarshalOptions{
		EmitUnpopulated: false,
	}
	value, err := marshaler.Marshal(bank)
	if err != nil {
		return err
	}
	write := runtime.StorageWrite{
		Collection: CollectionBank,
		Key:        bank.GetSenderId(),
		UserID:     bank.GetSenderId(),
		Value:      string(value),
	}
	writes := []*runtime.StorageWrite{&write}
	_, err = nk.StorageWrite(ctx, writes)
	if err != nil {
		logger.WithField("err", err).Error("Wallet update error.")
	}
	return err
}

func ReadBank(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger, userId string) (*pb.Bank, error) {
	unmarshaler := &protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}
	read := runtime.StorageRead{
		Collection: CollectionBank,
		Key:        userId,
		UserID:     userId,
	}
	reads := []*runtime.StorageRead{&read}
	objects, err := nk.StorageRead(ctx, reads)
	if err != nil {
		logger.WithField("err", err).Error("Wallet update error.")
		return nil, err
	}
	bank := &pb.Bank{}
	if len(objects) == 0 {
		bank.SenderId = userId
	} else {
		err = unmarshaler.Unmarshal([]byte(objects[0].Value), bank)
		if err != nil {
			logger.Error("Unmarshal Bank user %s, err: %s", userId, err.Error())
		}
	}

	return bank, err
}
