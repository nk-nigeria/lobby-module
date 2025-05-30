package entity

import (
	"context"
	"errors"
	"strconv"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgp-common/lib"
	pb "github.com/nakamaFramework/cgp-common/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	CollectionBank = "collection_bank"
)

func BankPushToSafe(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, unmarshaler *protojson.UnmarshalOptions, bank *pb.Bank) (*pb.Bank, error) {
	if bank.GetChipsInBank() <= 0 {
		return nil, errors.New("Chips push to safe must larger than zero")
	}

	wallets, err := ReadWalletUsers(ctx, nk, logger, bank.GetSenderId())
	if err != nil {
		return nil, err
	}
	if len(wallets) == 0 {
		return nil, errors.New("User not have wallet")
	}
	currentWallet := wallets[0]
	if currentWallet.Chips < bank.GetChipsInBank() {
		return nil, errors.New("User chips smaller than amout chip push to safe")
	}

	newBank := pb.Bank{
		SenderId:     bank.GetSenderId(),
		SenderSid:    bank.GetSenderSid(),
		RecipientId:  bank.GetRecipientId(),
		RecipientSid: bank.GetRecipientSid(),
		Chips:        -bank.GetChipsInBank(),
		ChipsInBank:  bank.GetChipsInBank(),
		Action:       pb.Bank_ACTION_PUSH_TO_SAFE,
	}
	// substract chip in wallet
	err = updateBank(ctx, nk, logger, &newBank)
	if err != nil {
		logger.Error("Update wallet when push to safe action error: %s", err.Error())
		return nil, err
	}
	// add chip in bank
	newBank.Chips = currentWallet.Chips + newBank.Chips
	newBank.ChipsInBank = currentWallet.ChipsInBank + newBank.ChipsInBank
	logger.Info("User id %s push %d to safe, wallet before: chips %d, bank: %d, wallets after: chips %d, bank %d",
		bank.GetSenderId(), bank.GetChipsInBank(), currentWallet.Chips,
		currentWallet.ChipsInBank, newBank.Chips,
		newBank.ChipsInBank)

	return &newBank, nil
}

func BankWithdraw(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, bank *pb.Bank) (*pb.Bank, error) {
	if bank.GetChips() <= 0 {
		return nil, errors.New("Amout chip withdraw must larger than zero")
	}

	wallets, err := ReadWalletUsers(ctx, nk, logger, bank.GetSenderId())
	if err != nil {
		return nil, err
	}
	if len(wallets) == 0 {
		return nil, errors.New("User not have wallet")
	}
	currentWallet := wallets[0]
	if currentWallet.ChipsInBank < bank.GetChips() {
		return nil, errors.New("User chips smaller than amout chip push to safe")
	}

	newBank := pb.Bank{
		SenderId:     bank.GetSenderId(),
		SenderSid:    bank.GetSenderSid(),
		RecipientId:  bank.GetRecipientId(),
		RecipientSid: bank.GetRecipientSid(),
		Chips:        bank.GetChips(),
		ChipsInBank:  -bank.GetChips(),
		Action:       pb.Bank_ACTION_WITHDRAW,
	}
	err = updateBank(ctx, nk, logger, &newBank)
	if err != nil {
		logger.Error("Withdraw to wallet user %s, amout chip %d, err: %s", bank.GetSenderId(), bank.GetAmountFee(), err.Error())
		return nil, err
	}

	newBank.Chips = currentWallet.Chips + newBank.Chips
	newBank.ChipsInBank = currentWallet.ChipsInBank + newBank.ChipsInBank
	logger.Info("User id %s push %d to safe, wallet before: chips %d, bank: %d, wallets after: chips %d, bank %d",
		bank.GetSenderId(), bank.GetChips(), currentWallet.Chips,
		currentWallet.ChipsInBank, newBank.Chips,
		newBank.ChipsInBank)

	return &newBank, nil
}

func BankSendGift(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, bank *pb.Bank) (*pb.Bank, error) {
	if bank.GetChips() <= 0 {
		return nil, errors.New("Chip send gift must larger than zero")
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

	if senderWallet.Chips < bank.GetChips() {
		logger.Error("Sender %s amout chip smaller than amout chip request send gift, chips in wallet: %d, request send gift: %d",
			bank.GetSenderId(), senderWallet.Chips, bank.GetChips())
		err = errors.New("Sender amout chip smaller than amout chip request send gift")
		return nil, err
	}

	senderNewWallet := lib.Wallet{}
	senderNewWallet.Chips = -bank.Chips - AbsInt64(bank.AmountFee)
	newSenderBank := pb.Bank{
		SenderId:     bank.SenderId,
		SenderSid:    bank.SenderSid,
		RecipientId:  bank.RecipientId,
		RecipientSid: bank.RecipientSid,
		// Chips:        senderNewWallet.Chips,
		// ChipsInBank: senderNewWallet.ChipsInBank,
		Chips:  senderNewWallet.Chips,
		Action: pb.Bank_ACTION_SEND_GIFT,
	}
	err = updateBank(ctx, nk, logger, &newSenderBank)
	if err != nil {
		data, _ := protojson.Marshal(&newSenderBank)
		logger.Error("Update wallet sender %s error: %s, data %s", bank.GetSenderId(), err.Error(), string(data))
		return nil, err
	}
	newSenderBank.Chips += senderWallet.Chips
	return &newSenderBank, nil
}

func BankHistory(logger runtime.Logger, nk runtime.NakamaModule, userID string, limit, offset int64) {
}

func updateBank(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger, bank *pb.Bank) error {
	metadata := make(map[string]interface{})
	metadata["bank_action"] = bank.GetAction().String()
	metadata["sender"] = strconv.FormatInt(bank.GetSenderSid(), 10)
	metadata["recv"] = strconv.FormatInt(bank.GetRecipientSid(), 10)
	metadata["action"] = WalletActionBankTopup

	wallet := lib.Wallet{
		Chips:       bank.GetChips(),
		ChipsInBank: bank.GetChipsInBank(),
	}
	userId := bank.GetSenderId()
	return AddChipWalletUser(ctx, nk, logger, userId, wallet, metadata)
}
