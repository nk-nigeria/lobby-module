package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
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

const (
	kReferRewardCollection = "refer-reward-collection"
	kReferRewardKey        = "refer-reward"
)

var ListReferReward = &pb.ListRewardReferTemplate{
	RewardRefers: make([]*pb.RewardReferTemplate, 0),
}

func InitReferUserReward(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) {
	objectIds := []*runtime.StorageRead{
		{
			Collection: kReferRewardCollection,
			Key:        kReferRewardKey,
		},
	}
	objects, err := nk.StorageRead(ctx, objectIds)
	if err != nil {
		logger.Error("Error when read list refer reward at init, error %s", err.Error())
	}

	// check list game has write in collection
	if len(objects) > 0 {
		logger.Info("List refer reward already write in collection")
		listReferReward := &pb.ListRewardReferTemplate{}
		err = conf.Unmarshaler.Unmarshal([]byte(objects[0].GetValue()), listReferReward)
		if err == nil {
			sort.Slice(listReferReward.RewardRefers, func(i, j int) bool {
				a := listReferReward.RewardRefers[i].Min
				b := listReferReward.RewardRefers[j].Min
				return a < b
			})
			ListReferReward = listReferReward
		}
		return
	}

	// writeObjects := []*runtime.StorageWrite{}
	listReferReward := pb.ListRewardReferTemplate{}
	oneK := int64(1000)
	oneM := 1000 * oneK
	listReferReward.RewardRefers = []*pb.RewardReferTemplate{
		{
			Min:  0,
			Max:  10 * oneM,
			Rate: 0.1,
		},
		{
			Min:  10 * oneM,
			Max:  30 * oneM,
			Rate: 0.2,
		},
		{
			Min:  30 * oneM,
			Max:  50 * oneM,
			Rate: 0.3,
		},
		{
			Min:  50 * oneM,
			Max:  100 * oneM,
			Rate: 0.4,
		},
		{
			Min:  100 * oneM,
			Max:  0 * oneM,
			Rate: 0.5,
		},
	}

	sort.Slice(listReferReward.RewardRefers, func(i, j int) bool {
		a := listReferReward.RewardRefers[i].Min
		b := listReferReward.RewardRefers[j].Min
		return a < b
	})
	ListReferReward = &listReferReward
	exChangedealsJson, err := conf.Marshaler.Marshal(&listReferReward)
	if err != nil {
		logger.Debug("Can not marshaler list refer reward for collection")
		return
	}

	writeObjects := []*runtime.StorageWrite{
		{
			Collection:      kReferRewardCollection,
			Key:             kReferRewardKey,
			Value:           string(exChangedealsJson),
			PermissionRead:  2,
			PermissionWrite: 0,
		},
	}

	if len(writeObjects) == 0 {
		logger.Debug("Can not generate list refer reward for collection")
		return
	}

	_, err = nk.StorageWrite(ctx, writeObjects)
	if err != nil {
		logger.Error("Write list refer reward collection error %s", err.Error())
	}
}

func RpcEstRewardThisWeek() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}
		beginWeek, endWeek := entity.RangeThisWeek()
		req := &entity.FeeGameListCursor{
			UserId: userID,
			From:   beginWeek.Unix(),
			To:     endWeek.Unix(),
		}
		reward, err := cgbdb.GetRewardRefer(ctx, logger, db, req)
		if reward.Id == 0 || reward.UpdateTimeUnix < time.Now().Add(-1*time.Hour).Unix() {
			reward, err = EstRewardByUserId(ctx, logger, db, nk, req)
		}
		if err != nil {
			return "", presenter.ErrInternalError
		}
		reward.ListRewards = ListReferReward.GetRewardRefers()
		reward.RemainTimeResetSec = endWeek.Unix() - time.Now().Unix()
		out, _ := conf.MarshalerDefault.Marshal(reward)
		// EstRewardThisWeek(ctx, logger, db, nk, userID)
		return string(out), nil
	}
}

func RpcRewardHistory() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}
		historyRewardRequest := &pb.HistoryRewardRequest{}
		if err := conf.Unmarshaler.Unmarshal([]byte(payload), historyRewardRequest); err != nil {
			logger.Error("Error when unmarshal payload %s ", err.Error())
			return "", presenter.ErrUnmarshal
		}
		if historyRewardRequest.Time > pb.HistoryRewardTime_HISTORY_REWARD_TIME_LAST_MONTH {
			historyRewardRequest.Time = pb.HistoryRewardTime_HISTORY_REWARD_TIME_THIS_WEEK
		}

		if historyRewardRequest.Time >= 0 {
			var from time.Time
			var to time.Time
			switch historyRewardRequest.Time {
			case pb.HistoryRewardTime_HISTORY_REWARD_TIME_THIS_WEEK:
				from, to = entity.RangeThisWeek()
			case pb.HistoryRewardTime_HISTORY_REWARD_TIME_LAST_WEEK:
				from, to = entity.RangeLastWeek()
			case pb.HistoryRewardTime_HISTORY_REWARD_TIME_THIS_MONTH:
				from, to = entity.RangeThisMonth()
			case pb.HistoryRewardTime_HISTORY_REWARD_TIME_LAST_MONTH:
				from, to = entity.RangeLastMonth()
			}
			historyRewardRequest.From = from.Unix()
			historyRewardRequest.To = to.Unix()
		}
		historyRewardRequest.UserId = userID
		ml, err := cgbdb.GetHistoryRewardRefer(ctx, logger, db, historyRewardRequest)
		if err != nil {
			logger.Error("GetHistoryRewardRefer user %s , from %d --> %d, err %s",
				userID, historyRewardRequest.GetFrom(), historyRewardRequest.GetTo(), err.Error())
			return "", err
		}
		summary := &pb.RewardRefer{
			UserId:   userID,
			FromUnix: historyRewardRequest.From,
			ToUnix:   historyRewardRequest.To,
		}
		for _, r := range ml {
			summary.EstReward += r.GetEstReward()
			summary.UserRefers = append(summary.UserRefers, r.GetUserRefers()...)
		}
		// count user refer last online < 60d
		listUserRefer, _ := cgbdb.ListUserInvitedByUserId(ctx, logger, db, userID)
		listUserReferId := make([]string, 0, len(listUserRefer))
		for _, u := range listUserRefer {
			listUserReferId = append(listUserReferId, u.UserInvitee)
		}
		listAccount, _ := nk.AccountsGetId(ctx, listUserReferId)
		for _, ac := range listAccount {
			var metadata map[string]interface{}
			if err := json.Unmarshal([]byte(ac.User.GetMetadata()), &metadata); err != nil {
				continue
			}
			lastOnlineUnix := entity.ToInt64(metadata["last_online_time_unix"], 0)
			if time.Unix(lastOnlineUnix, 0).Add(60 * 24 * time.Hour).After(time.Now()) {
				summary.TotalUserRefer++
			}
		}

		out, _ := conf.MarshalerDefault.Marshal(summary)
		return string(out), nil
	}
}

func EstRewardThisWeekByUserId(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, userID string) {
	beginWeek, endWeek := entity.RangeThisWeek()
	req := &entity.FeeGameListCursor{
		UserId: userID,
		From:   beginWeek.Unix(),
		To:     endWeek.Unix(),
	}
	EstRewardByUserId(ctx, logger, db, nk, req)
}

func EstRewardLastWeekByUserId(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, userID string) {
	beginWeek, endWeek := entity.RangeLastWeek()
	req := &entity.FeeGameListCursor{
		UserId: userID,
		From:   beginWeek.Unix(),
		To:     endWeek.Unix(),
	}
	EstRewardByUserId(ctx, logger, db, nk, req)
}

func EstRewardByUserId(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, req *entity.FeeGameListCursor) (*pb.RewardRefer, error) {
	sumFee, err := cgbdb.GetSumFeeByUserId(ctx, logger, db, req)
	if err != nil {
		return nil, presenter.ErrInternalError
	}
	userID := req.UserId
	rewardRefer := &pb.RewardRefer{
		UserId:      userID,
		WinAmt:      sumFee.Fee,
		ListRewards: ListReferReward.GetRewardRefers(),
	}
	for idx, r := range ListReferReward.RewardRefers {
		if rewardRefer.WinAmt >= r.Min {
			rewardRefer.EstRewardLv = int64(idx + 1)
			rewardRefer.EstRateReward = r.Rate
			continue
		}
		if rewardRefer.WinAmt < r.Min {
			break
		}
	}
	rewardRefer.UserRefers, err = EstRewardFromReferredUser(ctx, logger, db, nk, req)
	if err != nil {
		logger.Error("EstRewardFromReferredUser %s err %s", userID, err.Error())
		// return "",
		return nil, errors.New("Est reward from referred user error")
	}
	for _, r := range rewardRefer.GetUserRefers() {
		r.EstRewardLv = rewardRefer.EstRewardLv
		r.EstReward = int64(float32(r.WinAmt) * rewardRefer.EstRateReward)
		rewardRefer.EstReward += r.EstReward
	}

	rewardRefer.FromUnix = req.From
	rewardRefer.ToUnix = req.To
	_, err = cgbdb.AddOrUpdateIfExistRewardRefer(ctx, logger, db, rewardRefer)
	return rewardRefer, err
}

func EstRewardFromReferredUser(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, req *entity.FeeGameListCursor) ([]*pb.RewardRefer, error) {
	listReferUser, err := cgbdb.ListUserInvitedByUserId(ctx, logger, db, req.UserId)
	if err != nil {
		logger.Error("Get list user prefer by user %s err %s", req.UserId, err.Error())
		return nil, err
	}
	listUserPreferReward := make([]*pb.RewardRefer, 0, len(listReferUser))
	for _, preferUser := range listReferUser {
		sumFee, err := cgbdb.GetSumFeeByUserId(ctx, logger, db, &entity.FeeGameListCursor{
			UserId: preferUser.UserInvitee,
			From:   req.From,
			To:     req.To,
		})
		if err != nil {
			return nil, presenter.ErrInternalError
		}
		rewardRefer := &pb.RewardRefer{
			UserId:   preferUser.UserInvitee,
			WinAmt:   sumFee.Fee,
			FromUnix: req.From,
			ToUnix:   req.To,
		}
		listUserPreferReward = append(listUserPreferReward, rewardRefer)
	}
	return listUserPreferReward, nil
}

func EstRewardThisWeek(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) {
	_, endWeek := entity.RangeThisWeek()
	listUser, err := cgbdb.GetAllUserHasReferLeastOneUser(ctx, logger, db, &endWeek)
	if err != nil {
		return
	}
	if len(listUser) == 0 {
		return
	}
	for _, u := range listUser {
		EstRewardThisWeekByUserId(ctx, logger, db, nk, u.UserInvitor)
	}
}

// run when go to new week
// est all reward refer user
func EstRewardLastWeek(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) {
	_, endWeek := entity.RangeLastWeek()

	listUser, err := cgbdb.GetAllUserHasReferLeastOneUser(ctx, logger, db, &endWeek)
	if err != nil {
		return
	}
	if len(listUser) == 0 {
		return
	}
	for _, u := range listUser {
		EstRewardLastWeekByUserId(ctx, logger, db, nk, u.UserInvitor)
	}
}

func SendReferRewardToWallet(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) {
	EstRewardLastWeek(ctx, logger, db, nk)
	limit := int64(100)
	offset := int64(0)
	for {
		listReward, err := cgbdb.GetListRewardCompleteReferNotSendToWallet(ctx, logger, db, limit, offset)
		if err != nil {
			return
		}
		if len(listReward) == 0 {
			return
		}
		metadata := make(map[string]interface{})
		metadata["action"] = entity.WalletActionReferReward
		metadata["sender"] = constant.UUID_USER_SYSTEM

		wallet := lib.Wallet{}
		for _, reward := range listReward {
			wallet.UserId = reward.UserId
			wallet.Chips = reward.EstReward
			metadata["recv"] = reward.UserId
			err = entity.AddChipWalletUser(ctx, nk, logger, reward.UserId, wallet, metadata)
			if err != nil {
				logger.Error("AddChipWalletUser for reward refer error %s", err.Error())
				return
			}
			logger.Info("Send %d chips for refer reward id %d", wallet.Chips, reward.Id)
			cgbdb.UpdateRewardReferHasSendToWallet(ctx, logger, db, reward.Id)
		}
	}
}
