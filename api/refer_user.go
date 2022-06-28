package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgp-lobby-module/cgbdb"
	"github.com/ciaolink-game-platform/cgp-lobby-module/conf"
	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
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
			Collection:      kExchangeCollection,
			Key:             kExchangeKey,
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
		sumFee, err := cgbdb.GetSumFeeByUserId(ctx, logger, db, req)
		if err != nil {
			return "", presenter.ErrInternalError
		}
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
			return "", errors.New("Est reward from referred user error")
		}
		for _, r := range rewardRefer.GetUserRefers() {
			r.EstRewardLv = rewardRefer.EstRewardLv
			r.EstReward = int64(float32(r.WinAmt) * rewardRefer.EstRateReward)
			rewardRefer.EstReward += r.EstReward
		}

		rewardRefer.From = beginWeek.Unix()
		rewardRefer.To = endWeek.Unix()
		// cgbdb.AddNewHistoryRewardRefer(ctx, logger, db, rewardRefer)
		out, _ := conf.MarshalerDefault.Marshal(rewardRefer)
		return string(out), nil

	}
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
			UserId: preferUser.UserInvitee,
			WinAmt: sumFee.Fee,
			From:   req.From,
			To:     req.To,
		}
		listUserPreferReward = append(listUserPreferReward, rewardRefer)
	}
	return listUserPreferReward, nil
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
			UserId: userID,
			From:   historyRewardRequest.From,
			To:     historyRewardRequest.To,
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
