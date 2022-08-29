package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgp-lobby-module/cgbdb"
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
		_, err := entity.BankSendGift(ctx, logger, nk, bank)
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
		queryParms, ok := ctx.Value(runtime.RUNTIME_CTX_QUERY_PARAMS).(map[string][]string)
		if !ok {
			queryParms = make(map[string][]string)
		}

		limit := 100
		if arr := queryParms["limit"]; len(arr) > 0 {
			if l, err := strconv.Atoi(arr[0]); err == nil {
				limit = l
			}
		}
		cusor := ""
		if arr := queryParms["cusor"]; len(arr) > 0 {
			cusor = arr[0]
		}
		metaAction := make([]string, 0)
		logger.Info("%v", queryParms["meta_action"])
		if arr := queryParms["meta_action"]; len(arr) > 0 {
			list := strings.Split(arr[0], ",")
			for _, s := range list {
				s = strings.ToLower(strings.TrimSpace(s))
				// if _, exist := entity.MapWalletAction[s]; exist {
				metaAction = append(metaAction, s)
				// }
			}
		}
		if len(metaAction) == 0 {
			metaAction = append(metaAction, entity.WalletActionBankTopup.String())
		}
		userUuid, _ := uuid.FromString(userID)
		list, cusor, _, err := cgbdb.ListWalletLedger(ctx, logger, db, userUuid, metaAction, &limit, cusor)
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
