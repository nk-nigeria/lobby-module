package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	"github.com/nakamaFramework/cgb-lobby-module/cgbdb"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	"github.com/nakamaFramework/cgb-lobby-module/constant"
	"github.com/nakamaFramework/cgb-lobby-module/entity"

	"github.com/nakamaFramework/cgp-common/lib"
	pb "github.com/nakamaFramework/cgp-common/proto"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

func RpcAddClaimableFreeChip(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		// userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		// if !ok {
		// 	return "", errors.New("Missing user ID.")
		// }
		// if userID != entity.UUID_USER_SYSTEM {
		// 	return "", status.Error(codes.Unauthenticated, "UnAuthenticate")
		// }
		freeChip := &pb.FreeChip{}
		if err := unmarshaler.Unmarshal([]byte(payload), freeChip); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		if freeChip.Chips > constant.MaxChipAllowAdd {
			logger.Error("freeChip.Chips (%d) > MaxChipAllowAdd (%d)", freeChip.Chips, constant.MaxChipAllowAdd)
			return "", presenter.ErrInvalidInput
		}
		freeChip.SenderId = constant.UUID_USER_SYSTEM
		freeChip.Action = entity.WalletActionFreeChip.String()
		// check valid user
		// RecipientId is sId or uuid
		// uuid -> convert to sid
		var account *entity.Account
		var err error
		if userSid, _ := strconv.Atoi(freeChip.RecipientId); userSid > 0 {
			account, err = cgbdb.GetAccount(ctx, db, "", int64(userSid))
		} else {
			account, err = cgbdb.GetAccount(ctx, db, freeChip.RecipientId, 0)
		}
		if err != nil {
			logger.WithField("recipient id", freeChip.RecipientId).WithField("err", err).Error("get account failed")
			return "", presenter.ErrNoUserIdFound
		}
		freeChip.RecipientId = strconv.FormatInt(account.Sid, 10)
		err = cgbdb.AddClaimableFreeChip(ctx, logger, db, freeChip)
		if err != nil {
			return "", err
		}
		// noti := pb.Notification{
		// 	RecipientId: freeChip.RecipientId,
		// 	Type:        pb.TypeNotification_GIFT,
		// 	Title:       "Freechip",
		// 	Content:     "Freechip",
		// 	SenderId:    "",
		// 	Read:        false,
		// }
		// err = cgbdb.AddNotification(ctx, logger, db, nk, &noti)
		// if err != nil {
		// 	logger.Warn("Add freechip noti err %s, body %s",
		// 		err.Error(), freeChip.String())
		// }
		freeChipStr, _ := conf.MarshalerDefault.Marshal(freeChip)
		return string(freeChipStr), nil
	}
}

func RpcMarkAcceptListFreeChip(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if ok && userID != "" {
			return "", errors.New("umauth")
		}
		freeChip := &pb.FreeChip{}
		if payload != "" {
			if err := unmarshaler.Unmarshal([]byte(payload), freeChip); err != nil {
				logger.Error("Error when unmarshal payload", err.Error())
				return "", presenter.ErrUnmarshal
			}
		}
		var err error
		freeChip, err = cgbdb.MarkClaimableFreeChip(ctx, logger, db, freeChip)
		if err != nil {
			logger.WithField("err", err).Error("mark claim freechip failed")
			return "", presenter.ErrInternalError
		}
		if freeChip.GetClaimStaus() == pb.FreeChip_CLAIM_STATUS_WAIT_USER_CLAIM {
			userSid, _ := strconv.ParseInt(freeChip.RecipientId, 10, 64)
			account, err := cgbdb.GetAccount(ctx, db, "", userSid)
			if err != nil {
				logger.WithField("err", err).WithField("user sid", userSid).Error("get account failed")
			} else {
				noti := pb.Notification{
					RecipientId: account.User.Id,
					Type:        pb.TypeNotification_GIFT,
					Title:       "Freechip",
					Content:     "Freechip",
					SenderId:    "",
					Read:        false,
				}
				err = cgbdb.AddNotification(ctx, logger, db, nk, &noti)
				if err != nil {
					logger.Warn("Add freechip noti err %s, body %s",
						err.Error(), freeChip.String())
				}
			}
		}
		freeChipStr, _ := conf.MarshalerDefault.Marshal(freeChip)
		return string(freeChipStr), nil
	}
}

func RpcClaimFreeChip(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}
		req := &pb.FreeChip{}
		if err := unmarshaler.Unmarshal([]byte(payload), req); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		account, err := cgbdb.GetAccount(ctx, db, userID, 0)
		if err != nil {
			logger.WithField("user id", userID).WithField("err", err).Error("get account failed")
			return "", presenter.ErrNoUserIdFound
		}
		req.RecipientId = strconv.FormatInt(account.Sid, 10)
		freeChip, err := cgbdb.ClaimFreeChip(ctx, logger, db, req.Id, req.RecipientId)
		if err != nil {
			if freeChip == nil {
				freeChip = &pb.FreeChip{}
			}
			logger.WithField("user id", userID).WithField("freechip id", freeChip.Id).WithField("err", err).Error("claim free chip failed")
			return "", err
		}
		wallet := lib.Wallet{
			Chips:  freeChip.Chips,
			UserId: userID,
		}
		metadata := make(map[string]interface{})
		metadata["action"] = entity.WalletActionFreeChip
		if freeChip.GetAction() != "" {
			metadata["action"] = freeChip.GetAction()
		}
		metadata["sender"] = freeChip.GetSenderId()
		metadata["recv"] = userID

		err = entity.AddChipWalletUser(ctx, nk, logger, userID, wallet, metadata)
		if err != nil {
			logger.Error("Add chip user %s, after claim freechip error %s", userID, err.Error())
			return "", err
		}
		logger.Info("User %s claim %d from %s", userID, freeChip.Chips, freeChip.SenderId)
		freeChip.Claimable = false
		freeChipStr, _ := conf.MarshalerDefault.Marshal(freeChip)
		// emit event to doris
		{
			metadata["chips"] = strconv.Itoa(int(freeChip.Chips))
			metadata["user_id"] = userID
			payload, _ := json.Marshal(metadata)
			report := lib.NewReportGame(ctx)
			data, _, err := report.ReportEvent(ctx, "send-chip", userID, string(payload))
			logger.WithField("err", err).Info("Report event send-chip user %s , data %s, repsonse %s",
				userID, string(payload), string(data))

		}
		return string(freeChipStr), nil
	}
}

func RpcListClaimableFreeChip(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}
		freeChip := &pb.FreeChip{}
		{
			account, err := cgbdb.GetAccount(ctx, db, userID, 0)
			if err != nil {
				logger.WithField("user id", userID).WithField("err", err).Error("get account failed")
				return "", presenter.ErrNoUserIdFound
			}
			freeChip.RecipientId = strconv.FormatInt(account.Sid, 10)
		}
		list, err := cgbdb.GetFreeChipClaimableByUser(ctx, logger, db, freeChip.RecipientId)
		if err != nil {
			return "", err
		}
		listFreeChipStr, _ := conf.MarshalerDefault.Marshal(list)
		return string(listFreeChipStr), nil
	}
}

func RpcCheckClaimFreeChip(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}
		freeChip := &pb.FreeChip{}
		if err := unmarshaler.Unmarshal([]byte(payload), freeChip); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		{
			account, err := cgbdb.GetAccount(ctx, db, userID, 0)
			if err != nil {
				logger.WithField("user id", userID).WithField("err", err).Error("get account failed")
				return "", presenter.ErrNoUserIdFound
			}
			freeChip.RecipientId = strconv.FormatInt(account.Sid, 10)
		}
		var err error
		freeChip, err = cgbdb.GetFreeChipByIdByUser(ctx, logger, db, freeChip.Id, userID)
		if err != nil {
			return "", err
		}

		freeChipStr, _ := conf.MarshalerDefault.Marshal(freeChip)
		return string(freeChipStr), nil
	}
}

func RpcListFreeChip(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if ok && userID != "" {
			return "", errors.New("umauth")
		}
		freeChip := &pb.FreeChipRequest{}
		if payload != "" {
			if err := unmarshaler.Unmarshal([]byte(payload), freeChip); err != nil {
				logger.Error("Error when unmarshal payload", err.Error())
				return "", presenter.ErrUnmarshal
			}
		}
		// logger.Info("User id %s", userID)
		// {
		// 	account, err := cgbdb.GetAccount(ctx, db, userID, 0)
		// 	if err != nil {
		// 		logger.WithField("user id", userID).WithField("err", err).Error("get account failed")
		// 		return "", presenter.ErrNoUserIdFound
		// 	}
		// 	freeChip.UserId = strconv.FormatInt(account.Sid, 10)
		// }
		list, err := cgbdb.GetListFreeChip(ctx, logger, db,
			"", int(freeChip.GetClaimStaus()), freeChip.Limit, freeChip.Cusor)
		if err != nil {
			return "", err
		}
		listFreeChipStr, _ := conf.MarshalerDefault.Marshal(list)
		return string(listFreeChipStr), nil
	}
}
