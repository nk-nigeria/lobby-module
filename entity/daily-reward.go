package entity

type UserDailyReward struct {
	LastClaimUnix       int64 `json:"lastClaimUnix"`
	Streak              int64 `json:"streak"`
	CanClaimDailyReward bool  `json:"canClaimDailyReward"`
}

func (u *UserDailyReward) AddPlusStreak() {
	u.Streak++
	if u.Streak > 6 {
		u.Streak = 1
	}
}

func (u *UserDailyReward) GetStreak() int64 {
	if u.Streak <= 0 {
		u.Streak = 1
	}
	return u.Streak
}
