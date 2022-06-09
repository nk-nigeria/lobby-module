package api

import (
	"context"
	"database/sql"
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
		Dailies: []*pb.DayReward{
			{
				IndayReward: &pb.Reward{
					Chips:        2200,
					PercentBonus: 10.0,
					Streak:       1,
				},
				OnlineRewards: []*pb.Reward{
					{
						Chips:      100,
						SecsOnline: 3600,
						Streak:     1,
					},
					{
						Chips:      500,
						SecsOnline: 7200,
						Streak:     2,
					},
				},
			},
			{
				IndayReward: &pb.Reward{
					Chips:        4400,
					PercentBonus: 20.0,
					Streak:       2,
				},
				OnlineRewards: []*pb.Reward{
					{
						Chips:      200,
						SecsOnline: 3600,
						Streak:     1,
					},
					{
						Chips:      600,
						SecsOnline: 7200,
						Streak:     2,
					},
				},
			},
			{
				IndayReward: &pb.Reward{
					Chips:        8800,
					PercentBonus: 30.0,
					Streak:       3,
				},
				OnlineRewards: []*pb.Reward{
					{
						Chips:      300,
						SecsOnline: 3600,
						Streak:     1,
					},
					{
						Chips:      700,
						SecsOnline: 7200,
						Streak:     2,
					},
				},
			},
			{
				IndayReward: &pb.Reward{
					Chips:        17600,
					PercentBonus: 40.0,
					Streak:       4,
				},
				OnlineRewards: []*pb.Reward{
					{
						Chips:      400,
						SecsOnline: 3600,
						Streak:     1,
					},
					{
						Chips:      800,
						SecsOnline: 7200,
						Streak:     2,
					},
				},
			},
			{
				IndayReward: &pb.Reward{
					Chips:        35200,
					PercentBonus: 50.0,
					Streak:       5,
				},
				OnlineRewards: []*pb.Reward{
					{
						Chips:      500,
						SecsOnline: 3600,
						Streak:     1,
					},
					{
						Chips:      900,
						SecsOnline: 7200,
						Streak:     2,
					},
				},
			},
			{
				IndayReward: &pb.Reward{
					Chips:        70400,
					PercentBonus: 100.0,
					Streak:       6,
				},
				OnlineRewards: []*pb.Reward{
					{
						Chips:      600,
						SecsOnline: 3600,
						Streak:     1,
					},
					{
						Chips:      1000,
						SecsOnline: 7200,
						Streak:     2,
					},
				},
			},
		},
	}

	for _, d := range dailyRewardTemplate.Dailies {
		bonusChips := float32(d.GetIndayReward().GetChips()) * (d.GetIndayReward().GetPercentBonus() / 100.0)
		d.GetIndayReward().BonusChips = int64(bonusChips)
		d.GetIndayReward().TotalChips = d.GetIndayReward().GetChips() + d.GetIndayReward().GetBonusChips()
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

func canUserClaimDailyReward(d *entity.Reward) bool {
	t := time.Now()
	// return time.Unix(d.LastClaimUnix, 0).Before(t.Add(-5 * time.Second))
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	return time.Unix(d.LastClaimUnix, 0).Before(midnight)
}

func RpcCanClaimDailyReward() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		// userDayReward, _, err := GetLastDailyRewardObject(ctx, logger, nk)
		// if err != nil {
		// 	logger.Error("Error getting daily reward: %v", err)
		// 	return "", presenter.ErrInternalError
		// }

		// userDayReward.CanClaimDailyReward = canUserClaimDailyReward(*userDayReward)
		// call GetStreak for reset stream
		// if last claim not yesterday
		// _ = userDayReward.GetStreak()
		userDayReward, err := GetAndProcessDailyReward(ctx, logger, db, nk)
		if err != nil {
			logger.Error("GetAndProcessDailyReward error %s ", err.Error())
			return "", presenter.ErrInternalError
		}
		out, err := conf.MarshalerDefault.Marshal(&pb.DayReward{
			IndayReward:  userDayReward.IndayReward.Reward,
			OnlineReward: userDayReward.OnlineReward.Reward,
		})

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
		userDailyReward, dailyRewardObject, err := GetLastDailyRewardObject(ctx, logger, nk)
		if err != nil {
			logger.Error("Error getting daily reward: %v", err)
			return "", presenter.ErrInternalError
		}

		userDailyReward, err = GetAndProcessDailyReward(ctx, logger, db, nk)
		if err != nil {
			logger.Error("getDailyRewardByStreak error %s ", err.Error())
			return "", presenter.ErrInternalError
		}
		if !userDailyReward.IndayReward.CanClaim && !userDailyReward.OnlineReward.CanClaim {
			out, err := conf.MarshalerDefault.Marshal(&pb.DayReward{
				IndayReward:  userDailyReward.IndayReward.Reward,
				OnlineReward: userDailyReward.OnlineReward.Reward,
			})
			if err != nil {
				logger.Error("Marshal error: %v", err)
				return "", presenter.ErrUnmarshal
			}
			return string(out), nil
		}
		version := ""
		if dailyRewardObject != nil {
			version = dailyRewardObject.GetVersion()
		}

		// userDailyReward.Reward = nil
		userDailyReward.IndayReward.LastClaimUnix = time.Now().Unix()
		userDailyReward.OnlineReward.LastClaimUnix = time.Now().Unix()
		wallet := entity.Wallet{
			Chips: 0,
		}
		if userDailyReward.IndayReward.CanClaim {
			userDailyReward.IndayReward.NumClaim++
			userDailyReward.IndayReward.PlusStreak(6)
			wallet.Chips += userDailyReward.IndayReward.GetTotalChips()
		}
		if userDailyReward.OnlineReward.CanClaim {
			userDailyReward.OnlineReward.SecsOnline = userDailyReward.OnlineReward.SecsOnlineAfterClaim
			userDailyReward.OnlineReward.NumClaim++
			wallet.Chips += userDailyReward.OnlineReward.GetTotalChips()
		}
		err = SaveUserDailyReward(ctx, nk, logger, userDailyReward, version, userID)
		if err != nil {
			return "", err
		}

		metadata := make(map[string]interface{})
		metadata["action"] = "daily-reward"
		metadata["sender"] = constant.UUID_USER_SYSTEM
		metadata["recv"] = userID
		entity.AddChipWalletUser(ctx, nk, logger, userID, wallet, metadata)
		userDailyReward.IndayReward.CanClaim = false
		userDailyReward.OnlineReward.CanClaim = false
		out, err := conf.MarshalerDefault.Marshal(&pb.DayReward{
			IndayReward:  userDailyReward.IndayReward.Reward,
			OnlineReward: userDailyReward.OnlineReward.Reward,
		})
		if err != nil {
			logger.Error("Marshal error: %v", err)
			return "", presenter.ErrUnmarshal
		}
		logger.Debug("rpcClaimDailyReward resp: %v", string(out))
		return string(out), nil
	}
}

func GetLastDailyRewardObject(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) (*entity.DayReward, *api.StorageObject, error) {
	// d := &DayReward{
	// 	Reward: pb.Reward{
	// 		LastClaimUnix:        0,
	// 		Streak:               0,
	// 		NumClaim:             0,
	// 		SecsOnline:           0,
	// 		SecsOnlineAfterClaim: 0,
	// 	},
	// }
	d := &entity.DayReward{
		IndayReward:  entity.NewReward(),
		OnlineReward: entity.NewReward(),
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
	unmarshaler := conf.Unmarshaler
	g := &pb.DayReward{}
	if err := unmarshaler.Unmarshal([]byte(objects[0].GetValue()), g); err != nil {
		logger.Error("Unmarshal error: %v", err)
		return nil, nil, presenter.ErrMarshal
	}
	d.IndayReward = &entity.Reward{Reward: g.IndayReward}
	d.OnlineReward = &entity.Reward{Reward: g.OnlineReward}

	for _, r := range g.OnlineRewards {
		d.Onlinerewards = append(d.Onlinerewards, &entity.Reward{Reward: r})
	}
	if d.IndayReward == nil {
		d.IndayReward = entity.NewReward()
	}
	if d.OnlineReward == nil {
		d.OnlineReward = entity.NewReward()
	}

	return d, objects[0], nil

}

func GetAndProcessDailyReward(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (*entity.DayReward, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		return nil, presenter.ErrNoUserIdFound
	}

	userDailyReward, _, err := GetLastDailyRewardObject(ctx, logger, nk)
	if err != nil {
		logger.Error("Error getting daily reward: %v", err)
		return nil, presenter.ErrInternalError
	}

	userDailyReward.IndayReward.CanClaim = canUserClaimDailyReward(userDailyReward.IndayReward)

	account, _, err := GetProfileUser(ctx, nk, userID, nil)
	if err != nil {
		logger.Error("GetProfileUser error %s", err.Error())
		return nil, err
	}
	userDailyReward.OnlineReward.CalcSecsNotClaimReward(account.LastOnlineTimeUnix)

	userDailyReward, err = getDailyReward(userDailyReward)
	if err != nil {
		logger.Error("getDailyRewardByStreak error %s ", err.Error())
		return nil, presenter.ErrInternalError
	}

	return userDailyReward, nil

}
func getDailyReward(userDailyReward *entity.DayReward) (*entity.DayReward, error) {
	// call GetStreak for reset stream
	// if last claim not yesterday
	userDailyReward.IndayReward.ResetStreakIfNotClaimContinue()
	if userDailyReward.IndayReward.Streak > int64(len(DailyRewardTemplate.Dailies)) {
		return nil, errors.New("Out of index daily reward template")
	}
	userDailyReward.OnlineReward.SecsOnlineAfterClaim = userDailyReward.OnlineReward.SecsOnline
	r := DailyRewardTemplate.Dailies[userDailyReward.IndayReward.GetStreak()]
	reward := &pb.DayReward{
		IndayReward: &pb.Reward{
			Chips:         r.GetIndayReward().Chips,
			PercentBonus:  r.GetIndayReward().PercentBonus,
			BonusChips:    r.GetIndayReward().BonusChips,
			TotalChips:    r.GetIndayReward().TotalChips,
			Streak:        r.GetIndayReward().Streak,
			LastClaimUnix: r.GetIndayReward().LastClaimUnix,
			CanClaim:      r.GetIndayReward().CanClaim,
		},
		OnlineReward:  r.OnlineReward,
		OnlineRewards: r.OnlineRewards,
	}
	userDailyReward.IndayReward.Chips = r.GetIndayReward().Chips
	userDailyReward.IndayReward.PercentBonus = r.GetIndayReward().PercentBonus
	userDailyReward.IndayReward.BonusChips = r.GetIndayReward().BonusChips
	userDailyReward.IndayReward.TotalChips = r.GetIndayReward().GetTotalChips()
	var onlineReward *pb.Reward

	timesGetOnlineReward := userDailyReward.OnlineReward.NumClaim + 1
	for _, s := range reward.OnlineRewards {
		if s.Streak == timesGetOnlineReward {
			onlineReward = s
			break
		}
	}
	if onlineReward == nil && len(reward.OnlineRewards) > 0 {
		onlineReward = reward.OnlineRewards[len(reward.OnlineRewards)-1]
	}

	// reward.OnlineRewards = nil
	// userDailyReward = &entity.DayReward{
	// 	IndayReward: &entity.Reward{
	// 		Reward: r.IndayReward,
	// 	},
	// 	OnlineReward: &entity.Reward{
	// 		Reward: r.OnlineReward,
	// 	},
	// }
	// userDailyReward.OnlineReward = entity.NewReward()

	if onlineReward != nil && userDailyReward.OnlineReward.SecsOnline >= onlineReward.SecsOnline {
		ratio := userDailyReward.OnlineReward.SecsOnline / onlineReward.SecsOnline
		userDailyReward.OnlineReward.Chips = ratio * onlineReward.Chips
		userDailyReward.OnlineReward.SecsOnlineAfterClaim -= ratio * onlineReward.SecsOnline
		if userDailyReward.OnlineReward.SecsOnlineAfterClaim < 0 {
			userDailyReward.OnlineReward.SecsOnlineAfterClaim = 0
		}

	}
	if userDailyReward.OnlineReward.Chips > 0 {
		userDailyReward.OnlineReward.TotalChips = userDailyReward.OnlineReward.Chips
		userDailyReward.OnlineReward.CanClaim = true
	} else {
		userDailyReward.OnlineReward.CanClaim = false
	}
	return userDailyReward, nil
}

func SaveUserDailyReward(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger, userDailyReward *entity.DayReward, version, userID string) error {
	indayReward := userDailyReward.IndayReward.Reward
	onlineReward := userDailyReward.OnlineReward.Reward
	out, _ := conf.Marshaler.Marshal(&pb.DayReward{
		IndayReward: &pb.Reward{
			NumClaim:      indayReward.NumClaim,
			LastClaimUnix: indayReward.LastClaimUnix,
			Streak:        indayReward.GetStreak(),
		},
		OnlineReward: &pb.Reward{
			NumClaim:      onlineReward.NumClaim,
			LastClaimUnix: onlineReward.LastClaimUnix,
			SecsOnline:    onlineReward.SecsOnline,
		},
	})
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
