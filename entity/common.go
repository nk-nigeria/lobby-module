package entity

import (
	"context"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	ModuleName = "lobby"
)

type Games struct {
	List []Game `json:"games"`
}
type Game struct {
	Code    string `json:"code"`
	Layout  Layout `json:"layout"`
	LobbyId string `json:"lobbyId"`
	MinChip int    `json:"minChip"`
	Enable  bool   `json:"enable"`
}

type Layout struct {
	Col     int `json:"col"`
	Row     int `json:"row"`
	ColSpan int `json:"colSpan"`
	RowSpan int `json:"rowSpan"`
}

// type Bet struct {
// 	Amount  int  `json:"amount"`
// 	MinChip int  `json:"minChip,omitempty"`
// 	Enable  bool `json:"enable"`

// 	MarkUnit  int `json:"markUnit"`            // mức cược (chip)
// 	Xjoin     int `json:"xJoin,omitempty"`     // tài sản tối thiểu cho phép join bàn  (đơn vị: mức cược bàn)
// 	AGJoin    int `json:"agJoin,omitempty"`    // tài sản tối thiểu cho phép join bàn (đơn vị: chip)
// 	Xplaynow  int `json:"xPlaynow,omitempty"`  // tài sản tối thiểu dùng để xác định bàn chơi khi ấn Quick Start (đơn vị: mức cược bàn)
// 	AGPlaynow int `json:"agPlaynow,omitempty"` //tài sản tối thiểu dùng để xác định bàn chơi khi ấn Quick Start (đơn vị: chip)
// 	Xleave    int `json:"xLeave,omitempty"`    // tài sản tối thiểu để xác định đuổi khỏi bàn (đơn vị: mức cược bàn)
// 	AGLeave   int `json:"agLeave,omitempty"`   // tài sản tối thiểu để xác định đuổi khỏi bàn (đơn vị: chip)
// 	Xfee      int `json:"xFee,omitempty"`      // mức tài sản tối đa để áp dụng ""New Fee"" (đơn vị: mcb)
// 	AGFee     int `json:"agFee,omitempty"`     // mức tài sản tối đa để áp dụng ""New Fee"" (đơn vị: chip)
// 	NewFee    int `json:"newFee,omitempty"`    // mức tiền hồ áp dụng khi số chip mang vào =< Xfee hoặc AGFee
// }

// type Bets struct {
// 	List []Bet `json:"bets"`
// }

func LoadBets(code string, ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, unmarshaler *protojson.UnmarshalOptions) (*pb.Bets, error) {
	objectIds := []*runtime.StorageRead{
		{
			Collection: "bets",
			Key:        code,
		},
	}
	objects, err := nk.StorageRead(ctx, objectIds)
	bets := &pb.Bets{}
	if err != nil {
		logger.Error("Error when read list bet, error %s", err.Error())
		return bets, presenter.ErrMarshal
	}
	if len(objectIds) == 0 {
		logger.Warn("List bet in storage empty")
		return bets, nil
	}
	err = unmarshaler.Unmarshal([]byte(objects[0].GetValue()), bets)
	if err != nil {
		logger.Error("Error when unmarshal list bets, error %s", err.Error())
		return bets, presenter.ErrUnmarshal
	}
	return bets, nil
}
