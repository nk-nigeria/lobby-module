package api

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ciaolink-game-platform/cgb-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgb-lobby-module/cgbdb"
	"github.com/ciaolink-game-platform/cgb-lobby-module/conf"
	"github.com/ciaolink-game-platform/cgb-lobby-module/constant"
	"github.com/ciaolink-game-platform/cgb-lobby-module/entity"

	pb "github.com/ciaolink-game-platform/cgp-common/proto"

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
		freeChip.SenderId = constant.UUID_USER_SYSTEM
		freeChip.Action = entity.WalletActionFreeChip.String()
		err := cgbdb.AddClaimableFreeChip(ctx, logger, db, freeChip)
		if err != nil {
			return "", err
		}
		noti := pb.Notification{
			RecipientId: freeChip.RecipientId,
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
		freeChip := &pb.FreeChip{}
		if err := unmarshaler.Unmarshal([]byte(payload), freeChip); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		var err error
		freeChip, err = cgbdb.ClaimFreeChip(ctx, logger, db, freeChip.Id, userID)
		if err != nil {
			return "", err
		}
		wallet := entity.Wallet{
			Chips:  freeChip.Chips,
			UserId: userID,
		}
		metadata := make(map[string]interface{})
		metadata["action"] = entity.WalletActionFreeChip
		if freeChip.GetAction() != "" {
			metadata["action"] = freeChip.GetAction()
		}
		metadata["sender"] = freeChip.GetSenderId()
		metadata["recv"] = freeChip.GetRecipientId()

		err = entity.AddChipWalletUser(ctx, nk, logger, userID, wallet, metadata)
		if err != nil {
			logger.Error("Add chip user %s, after claim freechip error %s", userID, err.Error())
			return "", err
		}
		logger.Info("User %s claim %d from %s", userID, freeChip.Chips, freeChip.SenderId)
		freeChip.Claimable = false
		freeChipStr, _ := conf.MarshalerDefault.Marshal(freeChip)
		return string(freeChipStr), nil
	}
}

func RpcListClaimableFreeChip(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}
		list, err := cgbdb.GetFreeChipClaimableByUser(ctx, logger, db, userID)
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
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		// if ok && userID != "" {
		// 	return "", errors.New("UnAth.")
		// }
		freeChip := &pb.FreeChipRequest{}
		if payload != "" {
			if err := unmarshaler.Unmarshal([]byte(payload), freeChip); err != nil {
				logger.Error("Error when unmarshal payload", err.Error())
				return "", presenter.ErrUnmarshal
			}
		}
		logger.Info("User id %s", userID)
		list, err := cgbdb.GetListFreeChip(ctx, logger, db,
			freeChip.UserId, freeChip.Limit, freeChip.Cusor)
		if err != nil {
			return "", err
		}
		listFreeChipStr, _ := conf.MarshalerDefault.Marshal(list)
		return string(listFreeChipStr), nil
	}
}
