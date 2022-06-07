package entity

import (
	"time"

	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
)

type UserDailyReward struct {
	LastClaimUnix        int64      `json:"lastClaimUnix"`
	Streak               int64      `json:"streak"`
	CanClaimDailyReward  bool       `json:"canClaimDailyReward"`
	CanClaimOnlineReward bool       `json:"canClaimOnlineReward"`
	SecsOnlineNotClaim   int64      `json:"secOnlineNotClaim"`
	SecsOnlineAfterClaim int64      `json:"secOnlineAfterClaim"`
	TimesGetOnlineReward int64      `json:"timesGetOnlineReward"`
	Reward               *pb.Reward `json:"reward"`
}

func (u *UserDailyReward) PlusStreak() {
	u.Streak++
	if u.Streak > 6 {
		u.Streak = 1
	}
	u.resetStreakIfNotClaimContinue()
}

func (u *UserDailyReward) GetStreak() int64 {

	if u.Streak <= 0 {
		u.Streak = 1
	}
	if u.Streak > 1 {
		u.resetStreakIfNotClaimContinue()
	}
	return u.Streak
}

func (u *UserDailyReward) CalcSecsNotClaimReward(lastOnlineTimeUnix int64) {
	t := time.Now()
	ts := lastOnlineTimeUnix
	if ts < u.LastClaimUnix {
		ts = u.LastClaimUnix
	}
	if ts > t.Unix() {
		// invalid start timestamp
		return
	}
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	if !time.Unix(ts, 0).After(midnight) || !time.Unix(ts, 0).Before(midnight.Add(24*time.Hour)) {
		ts = midnight.Unix()
		// reset in midnight
		u.SecsOnlineNotClaim = 0
		u.TimesGetOnlineReward = 0
	}
	u.SecsOnlineNotClaim += t.Unix() - ts

}

func (u *UserDailyReward) resetStreakIfNotClaimContinue() {
	t := time.Now()
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	if u.LastClaimUnix < midnight.Unix() {
		// reset streak, not claim continue
		u.Streak = 1
	}
}

func (u *UserDailyReward) GetTotalChips() int64 {
	totalChip := int64(0)
	if u.Reward != nil {
		totalChip = u.Reward.GetTotalChips()
	}

	if u.Reward.OnlineReward != nil {
		totalChip += u.Reward.OnlineReward.Chips
	}
	return totalChip
}
