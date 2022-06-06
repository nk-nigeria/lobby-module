package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ciaolink-game-platform/cgp-lobby-module/constant"
	"github.com/ciaolink-game-platform/cgp-lobby-module/message_queue"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"time"

	"github.com/bwmarrin/snowflake"
	nkapi "github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api"
	"github.com/ciaolink-game-platform/cgp-lobby-module/conf"
	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	objectstorage "github.com/ciaolink-game-platform/cgp-lobby-module/object-storage"
)

const (
	rpcIdGameList    = "list_game"
	rpcIdFindMatch   = "find_match"
	rpcIdCreateMatch = "create_match"
	rpcIdQuickMatch  = "quick_match"

	rpcIdListBet = "list_bet"

	rpcUserChangePass = "user_change_pass"
	rpcLinkUsername   = "link_username"

	rpcGetProfile      = "get_profile"
	rpcUpdateProfile   = "update_profile"
	rpcUpdatePassword  = "update_password"
	rpcUpdateAvatar    = "update_avatar"
	rpcUpdateQuickChat = "update_quickchat"
	rpcGetQuickChat    = "get_quickchat"

	rpcPushToBank        = "push_to_bank"
	rpcWithDraw          = "with_draw"
	rpcBankSendGift      = "send_gift"
	rpcWalletTransaction = "wallet_transaction"

	//FreeChip
	rpcAddClaimableFreeChip  = "add_claimable_freechip"
	rpcClaimFreeChip         = "claim_freechip"
	rpcListClaimableFreeChip = "list_claimable_freechip"
	rpcCheckClaimFreeChip    = "check_claim_freechip"
	rpcListFreeChip          = "list_freechip"
	rpcListDeal              = "list_deal"
)

var (
	node *snowflake.Node
)

//noinspection GoUnusedExportedFunction
func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	initStart := time.Now()
	conf.Init()

	var err error
	node := conf.SnowlakeNode

	marshaler := conf.Marshaler
	unmarshaler := conf.Unmarshaler

	api.InitListGame(ctx, logger, nk)
	api.InitListBet(ctx, logger, nk)
	api.InitDeal(ctx, logger, nk, marshaler)
	api.InitLeaderBoard(ctx, logger, nk, unmarshaler)
	message_queue.InitNatsService(logger, constant.NastEndpoint)

	objStorage, err := InitObjectStorage(logger)
	if err != nil {
		objStorage = &objectstorage.EmptyStorage{}
	} else {
		objStorage.MakeBucket(entity.BucketAvatar)
	}

	if err := initializer.RegisterRpc(rpcIdGameList, api.RpcGameList(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcIdFindMatch, api.RpcFindMatch(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcIdCreateMatch, api.RpcCreateMatch(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcIdQuickMatch, api.RpcQuickMatch(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcIdListBet, api.RpcBetList(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcPushToBank, api.RpcPushToBank(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcWithDraw, api.RpcWithDraw(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcBankSendGift, api.RpcBankSendGift(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcWalletTransaction, api.RpcWalletTransaction(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcUserChangePass, api.RpcUserChangePass(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcLinkUsername, api.RpcLinkUsername(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcGetProfile, api.RpcGetProfile(marshaler, unmarshaler, objStorage)); err != nil {
		return err
	}

	// Free Chip
	if err := initializer.RegisterRpc(rpcAddClaimableFreeChip, api.RpcAddClaimableFreeChip(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcClaimFreeChip, api.RpcClaimFreeChip(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcListClaimableFreeChip, api.RpcListClaimableFreeChip(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcCheckClaimFreeChip, api.RpcCheckClaimFreeChip(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcListFreeChip, api.RpcListFreeChip(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcListDeal, api.RpcDealList(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterBeforeAuthenticateDevice(func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *nkapi.AuthenticateDeviceRequest) (*nkapi.AuthenticateDeviceRequest, error) {
		newID := node.Generate().Int64()
		if in.Username == "" {
			in.Username = fmt.Sprintf("%s.%d", entity.AutoPrefix, newID)
		}

		return in, nil
	}); err != nil {
		return err
	}

	if err := initializer.RegisterBeforeAuthenticateFacebook(func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *nkapi.AuthenticateFacebookRequest) (*nkapi.AuthenticateFacebookRequest, error) {
		if in.Username == "" {
			newID := node.Generate().Int64()
			in.Username = fmt.Sprintf("%s.%d", entity.AutoPrefixFacebook, newID)
		}
		return in, nil
	}); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(
		rpcUpdateProfile,
		api.RpcUpdateProfile(marshaler, unmarshaler, objStorage),
	); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(
		rpcUpdatePassword,
		api.RpcUpdatePassword(marshaler, unmarshaler),
	); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(
		rpcUpdateAvatar,
		api.RpcUploadAvatar(marshaler, unmarshaler, objStorage),
	); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(
		rpcUpdateQuickChat,
		api.RpcUpdateQuickChat(marshaler, unmarshaler),
	); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(
		rpcGetQuickChat,
		api.RpcGetQuickChat(marshaler, unmarshaler),
	); err != nil {
		return err
	}

	if err := api.RegisterSessionEvents(db, nk, initializer); err != nil {
		return err
	}

	api.RegisterValidatePurchase(db, nk, initializer)

	message_queue.RegisterHandler("leaderboard_add_score", func(data []byte) {
		leaderBoardRecord := &pb.LeaderBoardRecord{}
		err := unmarshaler.Unmarshal(data, leaderBoardRecord)
		if err != nil {
			logger.Error("leaderboard_add_score unmarshaler err %v data %v", err, string(data))
			return
		}
		api.UpdateScoreLeaderBoard(ctx, logger, nk, leaderBoardRecord)
	})
	message_queue.GetNatsService().RegisterAllSubject()
	// api.RegisterSessionEvents()
	logger.Info("Plugin loaded in '%d' msec.", time.Now().Sub(initStart).Milliseconds())
	return nil
}

const (
	MinioHost      = "103.226.250.195:9000"
	MinioKey       = "minio"
	MinioAccessKey = "minioadmin"
)

func InitObjectStorage(logger runtime.Logger) (objectstorage.ObjStorage, error) {
	w, err := objectstorage.NewMinioWrapper(MinioHost, MinioKey, MinioAccessKey, false)
	if err != nil {
		logger.Error("Init Object Storage Engine Minio error: %s", err.Error())
	}
	return w, err
}
