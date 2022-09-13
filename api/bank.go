package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgp-lobby-module/cgbdb"
	"github.com/ciaolink-game-platform/cgp-lobby-module/conf"
	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
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
		unmarshaler.Unmarshal([]byte(payload), bank)
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
		unmarshaler.Unmarshal([]byte(payload), bank)
		bank.SenderId = userID
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
			return "", presenter.ErrMarshal
		}
		unmarshaler.Unmarshal([]byte(payload), bank)
		bank.SenderId = userID
		// bank.AmountFee = 3
		_, err := entity.BankSendGift(ctx, logger, nk, bank)
		if err != nil {
			return "", err
		}
		freeChip := &pb.FreeChip{
			SenderId:    bank.SenderId,
			RecipientId: bank.RecipientId,
			Title:       "User send gift",
			Content:     "User send gift",
			Chips:       bank.ChipsInBank,
			Action:      entity.WalletActionUserGift.String(),
		}
		err = cgbdb.AddClaimableFreeChip(ctx, logger, db, freeChip)
		if err != nil {
			return "", err
		}
		return "", nil
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
