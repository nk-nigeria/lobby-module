package api

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"time"

	"github.com/heroiclabs/nakama-common/api"
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
				OnlineSec:   150,
				OnlineChip:  100,
				Streak:      1,
			},
			{
				BasicChips:  []int64{1001, 2001, 3001, 4001, 5001, 6001, 7001},
				PercenBonus: 20,
				OnlineSec:   300,
				OnlineChip:  200,
				Streak:      2,
			},
			{
				BasicChips:  []int64{1002, 2002, 3002, 4002, 5002, 6002, 7002},
				PercenBonus: 30,
				OnlineSec:   300,
				OnlineChip:  300,
				Streak:      3,
			},
			{
				BasicChips:  []int64{1003, 2003, 3003, 4003, 5003, 6003, 7003},
				PercenBonus: 40,
				OnlineSec:   300,
				OnlineChip:  400,
				Streak:      4,
			},
			{
				BasicChips:  []int64{1004, 2004, 3004, 4004, 5004, 6004, 7004},
				PercenBonus: 50,
				OnlineSec:   300,
				OnlineChip:  500,
				Streak:      5,
			},
			{
				BasicChips:  []int64{1005, 2005, 3005, 4005, 5005, 6005, 7005},
				PercenBonus: 100,
				OnlineSec:   300,
				OnlineChip:  600,
				Streak:      6,
			},
		},
	}
	sort.Slice(dailyRewardTemplate.RewardTemplates, func(i, j int) bool {
		a := dailyRewardTemplate.RewardTemplates[i]
		b := dailyRewardTemplate.RewardTemplates[j]
		return a.Streak < b.Streak
	})
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
		logger.WithField("err", err).Error("Write deals collection failed")
	}
}

func RpcDailyRewardTemplate() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		_, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}
		objectIds := []*runtime.StorageRead{
			{
				Collection: kDailyRewardTemplateCollection,
				Key:        kDailyRewardTemplateKey,
			},
		}
		objects, err := nk.StorageRead(ctx, objectIds)
		if err != nil {
			logger.Error("Error when read daily reward template at init, error %s", err.Error())
			return "", presenter.ErrInternalError
		}
		if len(objects) > 0 {
			return objects[0].GetValue(), nil
		}
		return "{}", nil
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
		reward, err := proccessDailyReward(ctx, logger, nk, db)
		if err != nil {
			logger.Error("proccessDailyReward error ", err.Error())
			return "", err
		}
		return RewardToString(reward, logger)
	}
}

func RpcClaimDailyReward() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		reward, err := proccessDailyReward(ctx, logger, nk, db)
		if err != nil {
			logger.Error("proccessDailyReward error ", err.Error())
			return "", err
		}
		if !reward.CanClaim {
			return RewardToString(reward, logger)
		}
		lastClaim, lastClaimObject, err := GetLastDailyRewardObject(ctx, logger, nk)
		if err != nil {
			logger.Error("GetLastDailyRewardObject error ", err.Error())
			return "", err
		}
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		savelastClaim := &pb.LastClaimReward{
			LastClaimUnix:  time.Now().Unix(),
			NextClaimUnix:  0,
			Streak:         reward.Streak,
			LastSpinNumber: 0,
			ReachMaxStreak: reward.ReachMaxStreak,
			NumClaim:       lastClaim.GetNumClaim() + 1,
		}
		version := ""
		if lastClaimObject != nil {
			version = lastClaimObject.GetVersion()
		}
		SaveLastClaimReward(ctx, nk, logger, savelastClaim, version, userID)
		wallet := lib.Wallet{
			Chips: reward.TotalChip,
		}

		metadata := make(map[string]interface{})
		metadata["action"] = entity.WalletActionDailyReward
		metadata["sender"] = constant.UUID_USER_SYSTEM
		metadata["recv"] = userID
		entity.AddChipWalletUser(ctx, nk, logger, userID, wallet, metadata)

		reward, _ = proccessDailyReward(ctx, logger, nk, db)

		reward.CanClaim = false
		return RewardToString(reward, logger)
	}
}

func proccessDailyReward(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, db *sql.DB) (*pb.Reward, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		return nil, presenter.ErrNoUserIdFound
	}
	profile, _, err := cgbdb.GetProfileUser(ctx, db, userID, nil)
	if err != nil {
		logger.Error("Get account %s error %s", userID, err.Error())
		return nil, presenter.ErrInternalError
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
	d.LastOnlineUnix = profile.GetLastOnlineTimeUnix()
	d.NumClaim = lastClaim.GetNumClaim()
	// streak on ui start at 1, on sv start at 0
	d.Streak = lastClaim.GetStreak() + 1
	if d.NumClaim >= int64(len(DailyRewardTemplate.RewardTemplates)) ||
		d.ReachMaxStreak {
		if !d.ReachMaxStreak {
			needSaveLastClaim = true
			d.ReachMaxStreak = true
		}
		d.NextClaimSec = 0
		if needSaveLastClaim {
			SaveLastClaimReward(ctx, nk, logger, lastClaim, version, userID)
		}
		return d, nil
	}
	if lastClaim.GetStreak() >= int64(len(DailyRewardTemplate.RewardTemplates)) {
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
	d.BonusChip = int64(float32(d.GetBasicChip()) * (d.GetPercentBonus() / 100.0))
	d.OnlineChip = rewardTemplate.GetOnlineChip()
	if lastClaim.NextClaimUnix == 0 {
		needSaveLastClaim = true
		timestampCountReward := entity.MaxIn64(d.LastClaimUnix, d.LastOnlineUnix)
		lastClaim.NextClaimUnix = timestampCountReward + rewardTemplate.OnlineSec
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
	d.TotalChip = d.BasicChip + d.BonusChip + d.OnlineChip
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
	data, err := conf.MarshalerDefault.Marshal(r)
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
