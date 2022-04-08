package entity

import pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"

type Bet struct {
	Amount  int  `json:"amount"`
	MinChip int  `json:"minChip,omitempty"`
	Enable  bool `json:"enable"`

	MarkUnit  int32   `json:"markUnit"`            // mức cược (chip)
	Xjoin     int     `json:"xJoin,omitempty"`     // tài sản tối thiểu cho phép join bàn  (đơn vị: mức cược bàn)
	AGJoin    int     `json:"agJoin,omitempty"`    // tài sản tối thiểu cho phép join bàn (đơn vị: chip)
	Xplaynow  int     `json:"xPlaynow,omitempty"`  // tài sản tối thiểu dùng để xác định bàn chơi khi ấn Quick Start (đơn vị: mức cược bàn)
	AGPlaynow int     `json:"agPlaynow,omitempty"` //tài sản tối thiểu dùng để xác định bàn chơi khi ấn Quick Start (đơn vị: chip)
	Xleave    int     `json:"xLeave,omitempty"`    // tài sản tối thiểu để xác định đuổi khỏi bàn (đơn vị: mức cược bàn)
	AGLeave   int     `json:"agLeave,omitempty"`   // tài sản tối thiểu để xác định đuổi khỏi bàn (đơn vị: chip)
	Xfee      int     `json:"xFee,omitempty"`      // mức tài sản tối đa để áp dụng ""New Fee"" (đơn vị: mcb)
	AGFee     int     `json:"agFee,omitempty"`     // mức tài sản tối đa để áp dụng ""New Fee"" (đơn vị: chip)
	NewFee    float32 `json:"newFee,omitempty"`    // mức tiền hồ áp dụng khi số chip mang vào =< Xfee hoặc AGFee
}

func (b Bet) ToPb() *pb.Bet {
	return &pb.Bet{
		Enable:   b.Enable,
		MarkUnit: b.MarkUnit,
	}
}

type ListBets struct {
	Bets []Bet `json:"bets"`
}
