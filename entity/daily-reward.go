package entity

import (
	"time"

	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
)

type DayReward struct {
	IndayReward   *Reward   `json:"indayReward"`
	OnlineReward  *Reward   `json:"onlineReward"`
	Onlinerewards []*Reward `json:"onlineRewards"`
}
type Reward struct {
	*pb.Reward
}

func NewReward() *Reward {
	r := Reward{
		Reward: &pb.Reward{},
	}
	return &r
}

func (u *Reward) PlusStreak(maxStreak int64) {
	u.Streak++
	if u.Streak > maxStreak {
		u.Streak = 1
	}
	u.resetStreakIfNotClaimContinue()
}

func (u *Reward) GetStreak() int64 {

	if u.Streak <= 0 {
		u.Streak = 1
	}
	if u.Streak > 1 {
		u.resetStreakIfNotClaimContinue()
	}
	return u.Streak
}

func (u *Reward) CalcSecsNotClaimReward(lastOnlineTimeUnix int64) {
	t := time.Now()
	ts := lastOnlineTimeUnix
	lastClaimUnix := u.GetLastClaimUnix()
	if ts < lastClaimUnix {
		ts = lastClaimUnix
	}
	if ts > t.Unix() {
		// invalid start timestamp
		return
	}
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)

	// last claim not today -->  first claim in day
	if time.Unix(lastClaimUnix, 0).Before(t) {
		u.SecsOnline = 0
		u.NumClaim = 0
	}
	if !time.Unix(ts, 0).After(midnight) || !time.Unix(ts, 0).Before(midnight.Add(24*time.Hour)) {
		ts = midnight.Unix()
		// reset in midnight
	}
	u.SecsOnline += t.Unix() - ts

}

func (u *Reward) resetStreakIfNotClaimContinue() {
	t := time.Now()
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	if u.LastClaimUnix < midnight.Unix() {
		// reset streak, not claim continue
		u.Streak = 1
	}
}
