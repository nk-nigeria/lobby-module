package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ciaolink-game-platform/cgb-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgb-lobby-module/cgbdb"
	"github.com/ciaolink-game-platform/cgb-lobby-module/conf"
	"github.com/ciaolink-game-platform/cgb-lobby-module/constant"
	"github.com/ciaolink-game-platform/cgb-lobby-module/entity"
	"github.com/ciaolink-game-platform/cgp-common/lib"
	pb "github.com/ciaolink-game-platform/cgp-common/proto"
	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

func RpcPushToBank(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		bank := &pb.Bank{}
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		if payload == "" {
			return "", presenter.ErrMarshal
		}
		err := unmarshaler.Unmarshal([]byte(payload), bank)
		if err != nil {
			logger.WithField("err", err).Error("unmarshal payload failed")
			return "", err
		}
		profile, _, err := cgbdb.GetProfileUser(ctx, db, userID, nil)
		if err != nil {
			logger.WithField("err", err).Error("get profile failed")
			return "", err
		}
		if profile.VipLevel < constant.MinLvAllowUseBank {
			return "", presenter.ErrFuncDisableByVipLv
		}
		bank.SenderId = userID
		newBank, err := entity.BankPushToSafe(ctx, logger, nk, unmarshaler, bank)
		if err != nil {
			return "", err
		}
		newBankJson, _ := marshaler.Marshal(newBank)
		return string(newBankJson), nil
	}
}

func RpcWithDraw(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		bank := &pb.Bank{}
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		if payload == "" {
			return "", presenter.ErrMarshal
		}
		err := unmarshaler.Unmarshal([]byte(payload), bank)
		if err != nil {
			logger.WithField("err", err).Error("unmarshal payload failed")
			return "", err
		}
		profile, _, err := cgbdb.GetProfileUser(ctx, db, userID, nil)
		if err != nil {
			logger.WithField("err", err).Error("get profile failed")
			return "", err
		}
		if profile.VipLevel < constant.MinLvAllowUseBank {
			return "", presenter.ErrFuncDisableByVipLv
		}
		bank.SenderId = userID
		bank.SenderSid = profile.GetUserSid()
		newBank, err := entity.BankWithdraw(ctx, logger, nk, bank)
		if err != nil {
			return "", err
		}
		newBankJson, _ := marshaler.Marshal(newBank)
		return string(newBankJson), nil
	}
}

func RpcBankSendGift(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		bank := &pb.Bank{}
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}
		if payload == "" {
			return "", presenter.ErrNoInputAllowed
		}
		err := unmarshaler.Unmarshal([]byte(payload), bank)
		if err != nil {
			return "", presenter.ErrUnmarshal
		}
		// check sender
		{
			account, err := cgbdb.GetAccount(ctx, db, userID, 0)
			if err != nil {
				return "", presenter.ErrUserNotFound
			}
			bank.SenderId = userID
			bank.SenderSid = account.Sid
		}
		// check recv
		{

			account, err := cgbdb.GetAccount(ctx, db, bank.GetRecipientId(), bank.GetRecipientSid())
			if err != nil {
				return "", presenter.ErrUserNotFound
			}
			bank.RecipientId = account.User.Id
			bank.RecipientSid = account.Sid
		}
		// bank.AmountFee = 3
		if bank.Chips == 0 {
			bank.Chips = bank.ChipsInBank
		}
		_, err = entity.BankSendGift(ctx, logger, nk, bank)
		if err != nil {
			return "", err
		}
		freeChip := &pb.FreeChip{
			SenderId:    bank.SenderId,
			RecipientId: bank.RecipientId,
			Title:       "User send gift",
			Content:     "User send gift",
			Chips:       bank.Chips,
			Action:      entity.WalletActionUserGift.String(),
		}
		err = cgbdb.AddClaimableFreeChip(ctx, logger, db, freeChip)
		if err != nil {
			return "", err
		}
		// emit event doris
		{
			report := lib.NewReportGame(ctx)
			metadata := make(map[string]any)
			metadata["action"] = entity.WalletActionUserGift
			metadata["sender"] = constant.UUID_USER_SYSTEM
			metadata["recv"] = userID
			metadata["chips"] = strconv.Itoa(int(bank.Chips))
			metadata["user_id"] = userID
			payload, _ := json.Marshal(metadata)
			report.ReportEvent(ctx, "send-chip", userID, string(payload))
		}
		// todo send noti
		noti := pb.Notification{
			RecipientId: freeChip.RecipientId,
			Type:        pb.TypeNotification_GIFT,
			Title:       "Gift",
			Content:     freeChip.GetContent(),
			SenderId:    bank.GetSenderId(),
			Read:        false,
		}
		err = cgbdb.AddNotification(ctx, logger, db, nk, &noti)
		if err != nil {
			logger.Warn("Add freechip noti err %s, body %s",
				err.Error(), freeChip.String())
		}
		jsonStr, _ := marshaler.Marshal(freeChip)
		return string(jsonStr), nil
	}
}

func RpcWalletTransaction(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}
		walletTransReq := &pb.WalletTransRequest{}
		if len(payload) > 0 {
			err := conf.Unmarshaler.Unmarshal([]byte(payload), walletTransReq)
			if err != nil {
				return "", presenter.ErrMarshal
			}
		}
		queryParms, ok := ctx.Value(runtime.RUNTIME_CTX_QUERY_PARAMS).(map[string][]string)
		if !ok {
			queryParms = make(map[string][]string)
		}

		limit := 100
		if arr := queryParms["limit"]; len(arr) > 0 {
			if l, err := strconv.Atoi(arr[0]); err == nil {
				limit = l
			}
		} else {
			limit = int(walletTransReq.GetLimit())
		}
		if limit <= 0 {
			limit = 10
		}
		cusor := ""
		if arr := queryParms["cusor"]; len(arr) > 0 {
			cusor = arr[0]
		} else {
			cusor = walletTransReq.GetCusor()
		}
		metaAction := make([]string, 0)
		{
			arr := queryParms["meta_action"]
			var list []string
			if len(arr) > 0 {
				list = strings.Split(arr[0], ",")
			} else {
				list = strings.Split(walletTransReq.GetMetaAction(), ",")
			}
			for _, s := range list {
				s = strings.ToLower(strings.TrimSpace(s))
				if len(s) > 0 {
					metaAction = append(metaAction, s)
				}
			}
		}

		if len(metaAction) == 0 {
			metaAction = append(metaAction, entity.WalletActionBankTopup.String())
		}
		metaBankAction := make([]string, 0)
		{
			arr := queryParms["meta_bank_action"]
			var list []string
			if len(arr) > 0 {
				list = strings.Split(arr[0], ",")
			} else {
				list = strings.Split(walletTransReq.GetMetaBankAction(), ",")
			}
			for _, s := range list {
				s = strings.TrimSpace(s)
				if len(s) > 0 {
					if num, err := strconv.Atoi(s); err == nil {
						metaBankAction = append(metaBankAction,
							pb.Bank_Action(num).String())
					} else {
						metaBankAction = append(metaBankAction, s)
					}
				}
			}
		}
		userUuid, _ := uuid.FromString(userID)
		list, cusor, _, err := cgbdb.ListWalletLedger(ctx, logger, db, userUuid, metaAction, metaBankAction, &limit, cusor)
		// list, cusor, err := nk.WalletLedgerList(ctx, userID, limit, cusor)
		if err != nil {
			logger.Error("WalletLedgerList  user: %s, error: %s", userID, err.Error())
			return "", err
		}
		// logger.Info("String return %s", str)
		walletTrans := entity.WalletTransaction{
			Transactions: list,
			Cusor:        cusor,
		}
		walletTransStr, err := json.Marshal(walletTrans)
		if err != nil {
			logger.Error("Marshal list wallet error %s", err.Error())
			return "", err
		}
		return string(walletTransStr), nil
	}
}
