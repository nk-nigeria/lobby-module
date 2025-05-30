package entity

import (
	pb "github.com/nakamaFramework/cgp-common/proto"
)

type Bet struct {
	Id             int64             `gorm:"column:id" json:"id,omitempty"`
	GameId         int               `gorm:"column:game_id" json:"gameId,omitempty"`
	Enable         bool              `gorm:"-" json:"enable,omitempty"`
	MarkUnit       int               `gorm:"column:mark_unit" json:"markUnit,omitempty"`  // mức cược (chip)
	Xjoin          float32           `gorm:"column:x_join" json:"xJoin,omitempty"`        // tài sản tối thiểu cho phép join bàn  (đơn vị: mức cược bàn)
	AGJoin         int               `gorm:"-" json:"agJoin,omitempty"`                   // tài sản tối thiểu cho phép join bàn (đơn vị: chip)
	Xplaynow       float32           `gorm:"column:x_play_now" json:"xPlaynow,omitempty"` // tài sản tối thiểu dùng để xác định bàn chơi khi ấn Quick Start (đơn vị: mức cược bàn)
	AGPlaynow      int               `gorm:"-" json:"agPlaynow,omitempty"`                //tài sản tối thiểu dùng để xác định bàn chơi khi ấn Quick Start (đơn vị: chip)
	Xleave         float32           `gorm:"column:x_leave" json:"xLeave,omitempty"`      // tài sản tối thiểu để xác định đuổi khỏi bàn (đơn vị: mức cược bàn)
	AGLeave        int               `gorm:"-" json:"agLeave,omitempty"`                  // tài sản tối thiểu để xác định đuổi khỏi bàn (đơn vị: chip)
	Xfee           float32           `gorm:"column:x_fee" json:"xFee,omitempty"`          // mức tài sản tối đa để áp dụng ""New Fee"" (đơn vị: mcb)
	AGFee          int               `gorm:"-" json:"agFee,omitempty"`                    // mức tài sản tối đa để áp dụng ""New Fee"" (đơn vị: chip)
	NewFee         float32           `gorm:"column:new_fee" json:"newFee,omitempty"`      // mức tiền hồ áp dụng khi số chip mang vào =< Xfee hoặc AGFee
	CountPlaying   int               `gorm:"-" json:"count_playing,omitempty"`
	MinVip         int               `gorm:"-" json:"min_vip,omitempty"`
	MaxVip         int               `gorm:"-" json:"max_vip,omitempty"`
	BetDisableType pb.BetDisableType `gorm:"-"`
}

func (b Bet) ToPb() *pb.Bet {
	return &pb.Bet{
		Id:             b.Id,
		Enable:         b.Enable,
		MarkUnit:       float32(b.MarkUnit),
		GameId:         int64(b.GameId),
		XJoin:          (b.Xjoin),
		AgJoin:         int64(b.AGJoin),
		XPlayNow:       (b.Xplaynow),
		AgPlayNow:      int64(b.AGPlaynow),
		XLeave:         (b.Xleave),
		AgLeave:        int64(b.AGLeave),
		XFee:           (b.Xfee),
		AgFee:          int64(b.AGFee),
		NewFee:         b.NewFee,
		CountPlaying:   int64(b.CountPlaying),
		MinVip:         int64(b.MinVip),
		MaxVip:         int64(b.MaxVip),
		BetDisableType: b.BetDisableType,
	}
}

func PbBetToBet(pb *pb.Bet) *Bet {
	if pb == nil {
		return &Bet{}
	}
	return &Bet{
		Id:             pb.Id,
		Enable:         pb.GetEnable(),
		MarkUnit:       int(pb.GetMarkUnit()),
		GameId:         int(pb.GameId),
		Xjoin:          (pb.XJoin),
		AGJoin:         int(pb.AgJoin),
		Xplaynow:       (pb.XPlayNow),
		AGPlaynow:      int(pb.AgPlayNow),
		Xleave:         (pb.XLeave),
		AGLeave:        int(pb.AgLeave),
		Xfee:           (pb.XFee),
		AGFee:          int(pb.AgFee),
		NewFee:         pb.NewFee,
		MinVip:         int(pb.MinVip),
		MaxVip:         int(pb.MaxVip),
		BetDisableType: pb.BetDisableType,
	}
}

type ListBets struct {
	Bets []Bet `json:"bets"`
}
