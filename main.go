package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"github.com/nakamaFramework/cgb-lobby-module/cgbdb"
	"github.com/nakamaFramework/cgp-common/define"
	"google.golang.org/protobuf/types/known/timestamppb"

	nkapi "github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"

	"github.com/nakamaFramework/cgb-lobby-module/api"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	"github.com/nakamaFramework/cgb-lobby-module/entity"
	objectstorage "github.com/nakamaFramework/cgb-lobby-module/object-storage"
)

const (
	rpcIdGameList    = "list_game"
	rpcGameAdd       = "add_game"
	rpcIdFindMatch   = "find_match"
	rpcIdCreateMatch = "create_match"
	rpcIdQuickMatch  = "quick_match"
	rpcIdInfoMatch   = "info_match"

	rpcIdListBet = "list_bet"
	// bet admin
	rpcAdminAddBetAddNew = "admin_bet_add"
	rpcAdminbetUpdate    = "admin_bet_update"
	rpcAdminbetDelete    = "admin_bet_delete"
	rpcAdminQueryBet     = "admin_bet"

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
	rpcMarkAcceptFreeChip    = "mark_accept_freechip"
	rpcListDeal              = "list_deal"
	rpcListExchangeDeal      = "list_exchange_deal"
	rpcExchangeAdd           = "exchange_add"
	rpcCancelExchange        = "exchange_cancel"
	rpcListExchange          = "list_exchange"
	rpcListExchangeLock      = "exchange_lock"
	rpcListExchangeById      = "exchange_by_id"
	rpcUpdataStatusExchange  = "exchange_update_status"

	rpcIdDailyRewardTemplate = "dailyrewardtemplate"
	rpcIdCanClaimDailyReward = "canclaimdailyreward"
	rpcIdClaimDailyReward    = "claimdailyreward"

	// UserGroup
	rpcIdListUserGroup   = "list_user_group"
	rpcIdAddUserGroup    = "add_user_group"
	rpcIdUpdateUserGroup = "update_user_group"
	rpcIdDeleteUserGroup = "delete_user_group"

	//giftcode
	rpcIdAddGiftCode    = "gift_code_add"
	rpcIdClaimGiftCode  = "gift_code_claim"
	rpcIdListGiftCode   = "gift_code_list"
	rpcIdDeleteGiftCode = "gift_code_delete"

	// Notification
	rpcIdListNotification      = "list_notification"
	rpcIdAddNotification       = "add_notification"
	rpcIdReadNotification      = "read_notification"
	rpcIdDeleteNotification    = "delete_notification"
	rpcIdReadAllNotification   = "read_all_notification"
	rpcIdDeleteAllNotification = "delete_all_notification"

	// InAppMessage
	rpcIdListInAppMessage   = "list_in_app_message"
	rpcIdAddInAppMessage    = "add_in_app_message"
	rpcIdUpdateInAppMessage = "update_in_app_message"
	rpcIdDeleteInAppMessage = "delete_in_app_message"

	rpcIdGetPreSignPush = "pre_sign_put"

	// refer user
	rpcRewardReferEstInWeek = "reward_refer_est_in_week"
	// refer user
	rpcRewardReferHistory = "reward_refer_history"

	// IAP
	rpcIAP = "iap"

	// leader board
	rpcLeaderBoardInfo = "leaderboard_info"

	rpcRuleLucky          = "rule_lucky"
	rpcRuleLuckyAdd       = "rule_lucky_add"
	rpcRuleLuckyUpdate    = "rule_lucky_update"
	rpcRuleLuckyDelete    = "rule_lucky_delete"
	rpcRuleLuckyEmitEvent = "rule_lucky_emit_event"
	// Jackpot
	rpcJackpot = "jackpot"
)

const (
	topicLeaderBoardAddScore = "leaderboard_add_score"
)

// noinspection GoUnusedExportedFunction
func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	defer func() {
		logger.Info("InitModule load done")
	}()
	initStart := time.Now()
	conf.Init()
	define.Init()
	var err error
	node := conf.SnowlakeNode

	marshaler := conf.Marshaler
	unmarshaler := conf.Unmarshaler
	if true {
		cgbdb.RunMigrations(ctx, logger, db)
	}

	api.InitListGame(ctx, logger, db, nk)
	api.InitDeal(ctx, logger, nk, marshaler)
	api.InitDailyRewardTemplate(ctx, logger, nk)
	api.InitLeaderBoard(ctx, logger, nk, unmarshaler)
	// message_queue.InitNatsService(logger, constant.NastEndpoint)
	api.InitExchangeList(ctx, logger, nk)
	api.InitReferUserReward(ctx, logger, nk)

	s := gocron.NewScheduler(time.Local)

	s.Every(1).Day().At("00:01").Do(func() {
		logger.Info("Start SendReferRewardToWallet")
		api.SendReferRewardToWallet(ctx, logger, db, nk)
	})
	s.StartAsync()

	objStorage, err := InitObjectStorage(logger)
	if err != nil {
		objStorage = &objectstorage.EmptyStorage{}
	} else {
		objStorage.MakeBucket(entity.BucketAvatar)
		objStorage.MakeBucket(entity.BucketBanners)
	}

	// initializer.RegisterRpc(rpcTestRemoteNode, api.RpcTestProxyNode(nk))
	if err := initializer.RegisterRpc(rpcGameAdd, api.RpcGameAdd(marshaler, unmarshaler)); err != nil {
		return err
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

	if err := initializer.RegisterRpc(rpcIdQuickMatch, api.RpcQuickMatch); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdInfoMatch, api.RpcInfoMatch(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcIdListBet, api.RpcBetList(conf.MarshalerDefault, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcAdminAddBetAddNew, api.RpcAdminAddBet(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcAdminbetUpdate, api.RpcAdminUpdateBet(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcAdminbetDelete, api.RpcAdminDeleteBet(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcAdminQueryBet, api.RpcAdminListBet(marshaler, unmarshaler)); err != nil {
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
	if err := initializer.RegisterRpc(rpcMarkAcceptFreeChip, api.RpcMarkAcceptListFreeChip(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcListDeal, api.RpcDealList(marshaler, unmarshaler)); err != nil {
		return err
	}

	// exchange
	if err := initializer.RegisterRpc(rpcListExchangeDeal, api.RpcExChangedealsList()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcExchangeAdd, api.RpcRequestNewExchange()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcCancelExchange, api.RpcRequestCancelExchange()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcListExchangeById, api.RpcGetExchange()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcListExchange, api.RpcGetAllExchange()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcListExchangeLock, api.RpcExchangeLock()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcUpdataStatusExchange, api.RpcExchangeUpdateStatus()); err != nil {
		return err
	}

	// daily reward
	if err := initializer.RegisterRpc(rpcIdCanClaimDailyReward,
		api.RpcCanClaimDailyReward()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdClaimDailyReward,
		api.RpcClaimDailyReward()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdDailyRewardTemplate,
		api.RpcDailyRewardTemplate()); err != nil {
		return err
	}

	// user group
	if err := initializer.RegisterRpc(rpcIdListUserGroup,
		api.RpcListUserGroup(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdAddUserGroup,
		api.RpcAddUserGroup(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdUpdateUserGroup,
		api.RpcUpdateUserGroup(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdDeleteUserGroup,
		api.RpcDeleteUserGroup(marshaler, unmarshaler)); err != nil {
		return err
	}

	// giftcode
	if err := initializer.RegisterRpc(rpcIdAddGiftCode,
		api.RpcAddGiftCode()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdClaimGiftCode,
		api.RpcClaimGiftCode()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdListGiftCode,
		api.RpcListGiftCode()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdDeleteGiftCode,
		api.RpcDeleteGiftCode()); err != nil {
		return err
	}
	//end gift code

	// Notification
	if err := initializer.RegisterRpc(rpcIdListNotification,
		api.RpcListNotification(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdAddNotification,
		api.RpcAddNotification(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdReadNotification,
		api.RpcReadNotification(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdDeleteNotification,
		api.RpcDeleteNotification(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdReadAllNotification,
		api.RpcReadAllNotification(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdDeleteAllNotification,
		api.RpcDeleteAllNotification(marshaler, unmarshaler)); err != nil {
		return err
	}

	// in app message
	if err := initializer.RegisterRpc(rpcIdListInAppMessage,
		api.RpcListInAppMessage(marshaler, unmarshaler, objStorage)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdAddInAppMessage,
		api.RpcAddInAppMessage(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdUpdateInAppMessage,
		api.RpcUpdateInAppMessage(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcIdDeleteInAppMessage,
		api.RpcDeleteInAppMessage(marshaler, unmarshaler)); err != nil {
		return err
	}

	// object storage
	if err := initializer.RegisterRpc(rpcIdGetPreSignPush,
		api.RpcGetPreSignPut(marshaler, unmarshaler, objStorage)); err != nil {
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
	if err := initializer.RegisterRpc(
		rpcRewardReferEstInWeek,
		api.RpcEstRewardThisWeek(),
	); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(
		rpcRewardReferHistory,
		api.RpcRewardHistory(),
	); err != nil {
		return err
	}

	// leaderboard
	if err := initializer.RegisterRpc(
		rpcLeaderBoardInfo,
		api.RpcLeaderboardInfo(),
	); err != nil {
		return err
	}

	// Rule lucky
	if err := initializer.RegisterRpc(rpcRuleLucky, api.RpcRuleLucky()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcRuleLuckyAdd, api.RpcRuleLuckyAdd()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcRuleLuckyUpdate, api.RpcRuleLuckyUpdate()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcRuleLuckyDelete, api.RpcRuleLuckyDelete()); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcRuleLuckyEmitEvent, api.RpcRuleLuckyEmitEvemt()); err != nil {
		return err
	}

	if err := api.RegisterSessionEvents(db, nk, initializer); err != nil {
		return err
	}

	api.RegisterValidatePurchase(db, nk, initializer)

	initializer.RegisterRpc(rpcIAP, api.RpcIAP())

	// custom nakama event
	initializer.RegisterEvent(api.CustomEventHandler(db))

	// message_queue.RegisterHandler(topicLeaderBoardAddScore, func(data []byte) {
	// 	leaderBoardRecord := &pb.LeaderBoardRecord{}
	// 	err := unmarshaler.Unmarshal(data, leaderBoardRecord)
	// 	if err != nil {
	// 		logger.Error("leaderboard_add_score unmarshaler err %v data %v", err, string(data))
	// 		return
	// 	}
	// 	api.UpdateScoreLeaderBoard(ctx, logger, nk, leaderBoardRecord)
	// })

	// message_queue.GetNatsService().RegisterAllSubject()
	cgbdb.AutoMigrate(db)

	// CreateAccountBot(ctx, db, logger)
	// api.RegisterSessionEvents()
	logger.Info("Plugin loaded in '%d' msec.", time.Since(initStart).Milliseconds())
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

func CreateAccountBot(ctx context.Context, db *sql.DB, logger runtime.Logger) {
	file, err := os.Open("indonesian_names.txt")
	if err != nil {
		logger.WithField("err", err).Error("open indonesian_names.txt error")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Read and process each line
	maxAccountCreated := 100000
	count := 0
	maxAccoutPerGame := 10000
	games := []define.GameName{
		define.GapleDomino,
		define.ChinesePoker,
		define.SicboName,
		define.BaccaratName,
		define.ColorGameName,
		define.BlackjackName,
		define.BandarqqName,
		define.DragontigerName,
	}
	curGameIdx := 0
	for scanner.Scan() {
		line := scanner.Text()
		// Process the line here, for example, print it
		// fmt.Println(line)
		user := &nkapi.Account{
			User: &nkapi.User{
				Id:          uuid.New().String(),
				Username:    line,
				DisplayName: line,
				Metadata:    "{\"bot\":\"true\"}",
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
			},
			VerifyTime: timestamppb.Now(),
			// DisableTime: timestamppb.Now(),
			DisableTime: nil,
			Email:       strings.ReplaceAll(line, "", "") + "@gmail.com",
		}
		err = cgbdb.CreateNewUser(ctx, db, user)
		if err != nil {
			logger.WithField("err", err).Error("create new user bot error")
		} else {
			cgbdb.CreateNewUserbot(ctx, db, user.User.Id, string(games[curGameIdx].String()))
		}

		count++
		if count%maxAccoutPerGame == 0 {
			curGameIdx++
		}
		if curGameIdx >= len(games) {
			break
		}
		if count >= maxAccountCreated {
			break
		}
	}

}
