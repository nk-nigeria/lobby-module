package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

var DailyRewardTemplate = &pb.DailyReward{}

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
		DailyRewardTemplate = &pb.DailyReward{}
		unmarshaler.Unmarshal([]byte(objects[0].GetValue()), DailyRewardTemplate)
	}
	dailyRewardTemplate := pb.DailyReward{
		Dailies: []*pb.Reward{
			{
				Chips:        2200,
				PercentBonus: 10.0,
				Streak:       1,
			},
			{
				Chips:        2200,
				PercentBonus: 20.0,
				Streak:       2,
			},
			{
				Chips:        2200,
				PercentBonus: 30.0,
				Streak:       3,
			},
			{
				Chips:        2200,
				PercentBonus: 40.0,
				Streak:       4,
			},
			{
				Chips:        2200,
				PercentBonus: 50.0,
				Streak:       5,
			},
			{
				Chips:        2200,
				PercentBonus: 100.0,
				Streak:       6,
			},
		},
	}

	for _, d := range dailyRewardTemplate.Dailies {
		bonusChips := float32(d.GetChips()) * (d.GetPercentBonus() / 100.0)
		d.BonusChips = int64(bonusChips)
		d.TotalChips = d.GetChips() + d.GetBonusChips()
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

func canUserClaimDailyReward(d entity.UserDailyReward) bool {
	t := time.Now()
	// return time.Unix(d.LastClaimUnix, 0).Before(t.Add(-5 * time.Second))
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	return time.Unix(d.LastClaimUnix, 0).Before(midnight)
}

func RpcCanClaimDailyReward() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userDailyReward, _, err := getLastDailyRewardObject(ctx, logger, nk)
		if err != nil {
			logger.Error("Error getting daily reward: %v", err)
			return "", presenter.ErrInternalError
		}

		userDailyReward.CanClaimDailyReward = canUserClaimDailyReward(*userDailyReward)

		out, err := json.Marshal(userDailyReward)
		if err != nil {
			logger.Error("Marshal error: %v", err)
			return "", presenter.ErrMarshal
		}

		logger.Debug("rpcCanClaimDailyReward resp: %v", string(out))
		return string(out), nil
	}
}

func RpcClaimDailyReward() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}
		userDailyReward, dailyRewardObject, err := getLastDailyRewardObject(ctx, logger, nk)
		if err != nil {
			logger.Error("Error getting daily reward: %v", err)
			return "", presenter.ErrInternalError
		}
		userDailyReward.CanClaimDailyReward = canUserClaimDailyReward(*userDailyReward)
		if !userDailyReward.CanClaimDailyReward {
			out, err := json.Marshal(userDailyReward)
			if err != nil {
				logger.Error("Marshal error: %v", err)
				return "", presenter.ErrMarshal
			}

			logger.Debug("RpcClaimDailyReward resp: %v", string(out))
			return string(out), nil
		}
		userDailyReward.LastClaimUnix = time.Now().Unix()
		userDailyReward.AddPlusStreak()
		rewardTemplate, err := getDailyRewardByStreak(userDailyReward.GetStreak())
		if err != nil {
			logger.Error("getDailyRewardByStreak error %s ", err.Error())
			return "", presenter.ErrInternalError
		}
		version := ""
		if dailyRewardObject != nil {
			version = dailyRewardObject.GetVersion()
		}
		object, err := json.Marshal(userDailyReward)
		_, err = nk.StorageWrite(ctx, []*runtime.StorageWrite{{
			Collection:      kDailyRewardCollection,
			Key:             kDailyRewardKey,
			PermissionRead:  1,
			PermissionWrite: 0, // No client write.
			Value:           string(object),
			Version:         version,
			UserID:          userID,
		}})
		if err != nil {
			logger.Error("StorageWrite error: %v", err)
			return "", presenter.ErrInternalError
		}

		wallet := entity.Wallet{
			Chips: rewardTemplate.GetTotalChips(),
		}
		metadata := make(map[string]interface{})
		metadata["action"] = "daily-reward"
		metadata["sender"] = constant.UUID_USER_SYSTEM
		metadata["recv"] = userID
		entity.AddChipWalletUser(ctx, nk, logger, userID, wallet, metadata)
		out, err := json.Marshal(userDailyReward)
		if err != nil {
			logger.Error("Marshal error: %v", err)
			return "", presenter.ErrUnmarshal
		}

		logger.Debug("rpcClaimDailyReward resp: %v", string(out))
		return string(out), nil
	}
}

func getLastDailyRewardObject(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) (*entity.UserDailyReward, *api.StorageObject, error) {
	d := &entity.UserDailyReward{
		LastClaimUnix: 0,
		Streak:        0,
	}
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
	if err := json.Unmarshal([]byte(objects[0].GetValue()), &d); err != nil {
		logger.Error("Unmarshal error: %v", err)
		return nil, nil, presenter.ErrMarshal
	}
	return d, objects[0], nil
}

func getDailyRewardByStreak(streak int64) (*pb.Reward, error) {
	if streak > int64(len(DailyRewardTemplate.Dailies)) {
		return nil, errors.New("Out of index daily reward template")
	}
	return DailyRewardTemplate.Dailies[streak-1], nil
}
