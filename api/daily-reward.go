package api

import (
	"context"
	"database/sql"
	"time"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgp-lobby-module/conf"
	"github.com/ciaolink-game-platform/cgp-lobby-module/constant"
	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	kDailyRewardTemplateCollection = "daily-reward-template-collection"
	kDailyRewardTemplateKey        = "daily-reward-template-key"

	kDailyRewardCollection = "daily-reward-collection"
	kDailyRewardKey        = "daily-reward-key"
)

var DailyRewardTemplate = &pb.DailyRewardTemplate{}

func InitDailyRewardTemplate(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) {
	objectIds := []*runtime.StorageRead{
		{
			Collection: kDailyRewardTemplateCollection,
			Key:        kDailyRewardTemplateKey,
		},
	}
	objects, err := nk.StorageRead(ctx, objectIds)
	if err != nil {
		logger.Error("Error when read daily reward template at init, error %s", err.Error())
	}
	if len(objects) > 0 {
		logger.Info("Daily reward template write in collection")
		unmarshaler := conf.Unmarshaler
		DailyRewardTemplate = &pb.DailyRewardTemplate{}
		unmarshaler.Unmarshal([]byte(objects[0].GetValue()), DailyRewardTemplate)
	}
	dailyRewardTemplate := pb.DailyRewardTemplate{
		RewardTemplates: []*pb.RewardTemplate{
			{
				BasicChips:  []int64{1000, 2000, 3000, 4000, 5000, 6000, 7000},
				PercenBonus: 10,
				OnlineSec:   3600,
				OnlineChip:  100,
			},
			{
				BasicChips:  []int64{1001, 2001, 3001, 4001, 5001, 6001, 7001},
				PercenBonus: 20,
				OnlineSec:   7200,
				OnlineChip:  200,
			},
			{
				BasicChips:  []int64{1002, 2002, 3002, 4002, 5002, 6002, 7002},
				PercenBonus: 30,
				OnlineSec:   7200,
				OnlineChip:  300,
			},
			{
				BasicChips:  []int64{1003, 2003, 3003, 4003, 5003, 6003, 7003},
				PercenBonus: 40,
				OnlineSec:   7200,
				OnlineChip:  400,
			},
			{
				BasicChips:  []int64{1004, 2004, 3004, 4004, 5004, 6004, 7004},
				PercenBonus: 50,
				OnlineSec:   7200,
				OnlineChip:  500,
			},
			{
				BasicChips:  []int64{1005, 2005, 3005, 4005, 5005, 6005, 7005},
				PercenBonus: 100,
				OnlineSec:   7200,
				OnlineChip:  600,
			},
		},
	}

	DailyRewardTemplate = &dailyRewardTemplate
	marshaler := conf.Marshaler
	dailyRewardTemplateJson, err := marshaler.Marshal(&dailyRewardTemplate)
	writeObjects := []*runtime.StorageWrite{
		{
			Collection:      kDailyRewardTemplateCollection,
			Key:             kDailyRewardTemplateKey,
			Value:           string(dailyRewardTemplateJson),
			PermissionRead:  2,
			PermissionWrite: 0,
		},
	}
	if len(writeObjects) == 0 {
		logger.Debug("Can not generate deals for collection")
		return
	}

	_, err = nk.StorageWrite(ctx, writeObjects)
	if err != nil {
		logger.Error("Write deals collection error %s", err.Error())
	}
}

func canUserClaimDailyReward(d *pb.Reward) bool {
	t := time.Now()
	// return time.Unix(d.LastClaimUnix, 0).Before(t.Add(-5 * time.Second))
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	return time.Unix(d.LastClaimUnix, 0).Before(midnight)
}

func RpcCanClaimDailyReward() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		reward, err := proccessDailyReward(ctx, logger, nk)
		if err != nil {
			logger.Error("proccessDailyReward error ", err.Error())
			return "", err
		}
		return RewardToString(reward, logger)
	}
}

func RpcClaimDailyReward() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		reward, err := proccessDailyReward(ctx, logger, nk)
		if err != nil {
			logger.Error("proccessDailyReward error ", err.Error())
			return "", err
		}
		if !reward.CanClaim {
			return RewardToString(reward, logger)
		}
		_, lastClaimObject, err := GetLastDailyRewardObject(ctx, logger, nk)
		if err != nil {
			logger.Error("GetLastDailyRewardObject error ", err.Error())
			return "", err
		}
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		lastClaim := &pb.LastClaimReward{
			LastClaimUnix:  time.Now().Unix(),
			NextClaimUnix:  0,
			Streak:         reward.Streak,
			LastSpinNumber: 0,
			ReachMaxStreak: reward.ReachMaxStreak,
		}
		version := ""
		if lastClaimObject != nil {
			version = lastClaimObject.GetVersion()
		}
		SaveLastClaimReward(ctx, nk, logger, lastClaim, version, userID)
		wallet := entity.Wallet{
			Chips: reward.TotalChips,
		}

		metadata := make(map[string]interface{})
		metadata["action"] = "daily-reward"
		metadata["sender"] = constant.UUID_USER_SYSTEM
		metadata["recv"] = userID
		entity.AddChipWalletUser(ctx, nk, logger, userID, wallet, metadata)

		reward, _ = proccessDailyReward(ctx, logger, nk)

		reward.CanClaim = false
		return RewardToString(reward, logger)
	}
}

func proccessDailyReward(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) (*pb.Reward, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		return nil, presenter.ErrNoUserIdFound
	}
	lastClaim, lastClaimObject, err := GetLastDailyRewardObject(ctx, logger, nk)

	if err != nil {
		logger.Error("Error getting daily reward: %v", err)
		return nil, presenter.ErrInternalError
	}
	version := ""
	if lastClaimObject != nil {
		version = lastClaimObject.GetVersion()
	}
	t := time.Now()
	midnightUnix := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Unix()
	nextMidnightUnix := midnightUnix + 86400

	needSaveLastClaim := false
	if lastClaim.LastClaimUnix < midnightUnix {
		lastClaim.LastClaimUnix = midnightUnix
		needSaveLastClaim = true
		lastClaim = &pb.LastClaimReward{
			LastClaimUnix: midnightUnix,
		}
	}

	d := &pb.Reward{}
	d.LastClaimUnix = lastClaim.GetLastClaimUnix()
	d.NextClaimUnix = lastClaim.GetNextClaimUnix()
	d.ReachMaxStreak = lastClaim.ReachMaxStreak
	// streak on ui start at 1, on sv start at 0
	d.Streak = lastClaim.GetStreak() + 1
	if d.ReachMaxStreak || lastClaim.GetStreak() >= int64(len(DailyRewardTemplate.RewardTemplates)) {
		d.NextClaimSec = 0
		d.ReachMaxStreak = true
		lastClaim.ReachMaxStreak = d.ReachMaxStreak
		SaveLastClaimReward(ctx, nk, logger, lastClaim, version, userID)
		return d, nil
	}
	rewardTemplate := DailyRewardTemplate.RewardTemplates[lastClaim.GetStreak()]
	if lastClaim.LastSpinNumber > 0 {
		d.BasicChip = rewardTemplate.GetBasicChips()[lastClaim.LastSpinNumber-1]
	} else {
		randonIdx := entity.RandomInt(1, len(rewardTemplate.BasicChips))
		d.BasicChip = rewardTemplate.GetBasicChips()[randonIdx-1]
		lastClaim.LastSpinNumber = int64(randonIdx)
		needSaveLastClaim = true
	}
	d.LastSpinNumber = lastClaim.LastSpinNumber
	d.PercentBonus = rewardTemplate.GetPercenBonus()
	d.BonusChips = int64(float32(d.GetBasicChip()) * (d.GetPercentBonus() / 100.0))
	d.OnlineChip = rewardTemplate.GetOnlineChip()
	if lastClaim.NextClaimUnix == 0 {
		needSaveLastClaim = true
		lastClaim.NextClaimUnix = lastClaim.LastClaimUnix + rewardTemplate.OnlineSec
		if lastClaim.NextClaimUnix >= nextMidnightUnix {
			d.NextClaimSec = 0
			d.ReachMaxStreak = true
			lastClaim.ReachMaxStreak = d.ReachMaxStreak
		}
		d.NextClaimUnix = lastClaim.NextClaimUnix
	}
	if needSaveLastClaim {
		SaveLastClaimReward(ctx, nk, logger, lastClaim, version, userID)
	}
	if !d.ReachMaxStreak {
		//d.Streak++ // stream o ui start at 1, on sv start at 0
		if d.NextClaimUnix < time.Now().Unix() {
			d.CanClaim = true
			d.NextClaimSec = 0
		} else {
			d.CanClaim = false
			d.NextClaimSec = d.GetNextClaimUnix() - time.Now().Unix()
		}
	}
	d.TotalChips = d.BasicChip + d.BonusChips + d.OnlineChip
	return d, nil
}

func GetLastDailyRewardObject(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) (*pb.LastClaimReward, *api.StorageObject, error) {

	d := &pb.LastClaimReward{}
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		return d, nil, presenter.ErrNoUserIdFound
	}

	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{{
		Collection: kDailyRewardCollection,
		Key:        kDailyRewardKey,
		UserID:     userID,
	}})
	if err != nil {
		logger.Error("StorageRead error: %v", err)
		return d, nil, presenter.ErrInternalError
	}
	if len(objects) == 0 {
		return d, nil, nil
	}
	unmarshaler := conf.Unmarshaler

	if err := unmarshaler.Unmarshal([]byte(objects[0].GetValue()), d); err != nil {
		logger.Error("Unmarshal error: %v", err)
		return nil, nil, presenter.ErrMarshal
	}
	return d, objects[0], nil
}

func RewardToString(r *pb.Reward, logger runtime.Logger) (string, error) {
	data, err := conf.Marshaler.Marshal(r)
	if err != nil {
		logger.Error("RewardToJson error %s", err.Error())
		return "", presenter.ErrMarshal
	}
	return string(data), nil
}

func SaveLastClaimReward(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger, lastClaim *pb.LastClaimReward, version, userID string) error {
	out, _ := conf.Marshaler.Marshal(lastClaim)
	_, err := nk.StorageWrite(ctx, []*runtime.StorageWrite{
		{
			Collection:      kDailyRewardCollection,
			Key:             kDailyRewardKey,
			PermissionRead:  1,
			PermissionWrite: 0, // No client write.
			Value:           string(out),
			Version:         version,
			UserID:          userID,
		},
	})
	if err != nil {
		logger.Error("StorageWrite error: %v", err)
		return presenter.ErrInternalError
	}
	return nil
}
