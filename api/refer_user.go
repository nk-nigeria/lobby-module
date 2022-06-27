package api

import (
	"context"
	"database/sql"
	"errors"
	"sort"

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
		beginWeek, endWeek := entity.RangeWeekFromNow()
		req := &entity.FeeGameListCursor{
			UserId: userID,
			From:   beginWeek.Unix(),
			To:     endWeek.Unix() - 1, // -1 sec
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
		out, _ := conf.MarshalerDefault.Marshal(rewardRefer)
		cgbdb.AddNewHistoryRewardRefer(ctx, logger, db, rewardRefer)
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
		}
		listUserPreferReward = append(listUserPreferReward, rewardRefer)
	}
	return listUserPreferReward, nil
}
