package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	"github.com/nakamaFramework/cgb-lobby-module/cgbdb"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	"github.com/nakamaFramework/cgb-lobby-module/constant"
	"github.com/nakamaFramework/cgb-lobby-module/entity"
	"github.com/nakamaFramework/cgp-common/lib"
	pb "github.com/nakamaFramework/cgp-common/proto"
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
		profile, _, err := cgbdb.GetProfileUser(ctx, db, userID, nil)
		if err != nil {
			logger.Error("Error when get user %s, err %s", userID, err.Error())
			return "", presenter.ErrInternalError
		}
		giftCode.UserId = userID
		dbGiftCode, err := cgbdb.ClaimGiftCode(ctx, logger, db, giftCode, profile.GetVipLevel())
		if err != nil {
			logger.Error("ClaimGiftCode error %s", err.Error())
			return "", err
		}
		if dbGiftCode.ErrCode == 0 {
			wallet := lib.Wallet{
				Chips: dbGiftCode.Value,
			}
			metadata := make(map[string]interface{})
			metadata["action"] = entity.WalletActionGiftCode
			metadata["sender"] = constant.UUID_USER_SYSTEM
			metadata["recv"] = userID
			// convert int64 to string because missing value when save to wallet metadata
			// metadata save int64 as float64 cause missing value
			metadata["g_id"] = strconv.FormatInt(dbGiftCode.Id, 10)
			err = entity.AddChipWalletUser(ctx, nk, logger, userID, wallet, metadata)
			if err != nil {
				logger.Error("Update wallet chip by claim giftcode %s error %s", giftCode.GetCode(), err.Error())
			}
			// emit event doris
			{
				report := lib.NewReportGame(ctx)
				metadata["chips"] = strconv.Itoa(int(wallet.Chips))
				metadata["user_id"] = userID
				payload, _ := json.Marshal(metadata)
				report.ReportEvent(ctx, "send-chip", userID, string(payload))
			}
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
