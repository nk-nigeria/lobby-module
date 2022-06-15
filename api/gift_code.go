package api

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgp-lobby-module/cgbdb"
	"github.com/ciaolink-game-platform/cgp-lobby-module/conf"
	"github.com/ciaolink-game-platform/cgp-lobby-module/constant"
	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
)

func RpcAddGiftCode() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unauth")
		}
		giftCode := &pb.GiftCode{}
		if err := conf.Unmarshaler.Unmarshal([]byte(payload), giftCode); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		if giftCode.Code == "" || giftCode.Value < 0 || giftCode.NMax <= 0 ||
			giftCode.StartTimeUnix <= 0 || giftCode.EndTimeUnix <= time.Now().Unix() {
			logger.Error("Invalid payload")
			return "", presenter.ErrUnmarshal
		}
		dbGiftCode, err := cgbdb.AddNewGiftCode(ctx, logger, db, giftCode)
		if err != nil {
			logger.Error("AddNewGiftCode error %s", err.Error())
			return "", err
		}
		out, _ := conf.MarshalerDefault.Marshal(dbGiftCode)
		return string(out), nil
	}
}

func RpcClaimGiftCode() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}
		giftCode := &pb.GiftCode{}
		if err := conf.Unmarshaler.Unmarshal([]byte(payload), giftCode); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		if giftCode.Code == "" {
			logger.Error("Invalid payload")
			return "", presenter.ErrUnmarshal
		}
		giftCode.UserId = userID
		dbGiftCode, err := cgbdb.ClaimGiftCode(ctx, logger, db, giftCode)
		if err != nil {
			logger.Error("ClaimGiftCode error %s", err.Error())
			return "", err
		}
		wallet := entity.Wallet{
			Chips: dbGiftCode.Value,
		}
		metadata := make(map[string]interface{})
		metadata["action"] = "gift_code"
		metadata["sender"] = constant.UUID_USER_SYSTEM
		metadata["recv"] = userID
		metadata["g_id"] = dbGiftCode.Id
		err = entity.AddChipWalletUser(ctx, nk, logger, userID, wallet, metadata)
		if err != nil {
			logger.Error("Update wallet chip by claim giftcode %s error %s", giftCode.GetCode(), err.Error())
		}
		out, _ := conf.MarshalerDefault.Marshal(dbGiftCode)
		return string(out), nil
	}
}

func RpcListGiftCode() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unauth")
		}
		ml, err := cgbdb.GetListGiftCode(ctx, logger, db, nil)
		if err != nil {
			logger.Error("GetListGiftCode error %s", err.Error())
		}
		out, _ := conf.MarshalerDefault.Marshal(&ml)
		return string(out), nil
	}

}

func RpcDeleteGiftCode() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unauth")
		}
		giftCode := &pb.GiftCode{}
		if err := conf.Unmarshaler.Unmarshal([]byte(payload), giftCode); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		if giftCode.Code == "" {
			logger.Error("Invalid payload")
			return "", presenter.ErrUnmarshal
		}
		dbGiftCode, err := cgbdb.DeletedGiftCode(ctx, logger, db, giftCode)
		if err != nil {
			logger.Error("DeletedGiftCode %s error %s", giftCode.GetCode(), err.Error())
			return "", err
		}
		out, _ := conf.MarshalerDefault.Marshal(dbGiftCode)
		return string(out), nil
	}
}
