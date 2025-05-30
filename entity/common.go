package entity

import (
	"fmt"
	"strconv"

	pb "github.com/nakamaFramework/cgp-common/proto"
)

const (
	ModuleName         = "lobby"
	AutoPrefix         = "CGPD"
	AutoPrefixFacebook = "CGPF"
)

const (
	MIN_LENGTH_PASSWORD = 6
)

func init() {
	MapWalletAction := make(map[WalletAction]bool, 0)
	MapWalletAction[WalletActionBankTopup] = true
	MapWalletAction[WalletActionDailyReward] = true
	MapWalletAction[WalletActionFreeChip] = true
	MapWalletAction[WalletActionGiftCode] = true
	MapWalletAction[WalletActionIAPTopUp] = true
	MapWalletAction[WalletActionReferReward] = true
}

type CustomUser struct {
	Id       string
	UserId   string
	UserName string
}

type Games struct {
	List []Game `json:"games"`
}

func (gs Games) ToPB() []*pb.Game {
	pbGames := make([]*pb.Game, 0)
	for _, g := range gs.List {
		pbGames = append(pbGames, &pb.Game{
			Code: g.Code,
			// Layout: &pb.Layout{
			// 	Col:     g.Layout.Col,
			// 	Row:     g.Layout.Row,
			// 	ColSpan: g.Layout.ColSpan,
			// 	RowSpan: g.Layout.RowSpan,
			// },
			LobbyId: g.LobbyId,
		})
	}

	return pbGames
}

type Game struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `json:"code"`
	// Layout      Layout  `gorm:"-" json:"layout"`
	LobbyId string `gorm:"-" json:"lobbyId"`
	// MinChip     int     `gorm:"-" json:"minChip"`
	// Enable      bool    `gorm:"-" json:"enable"`
	// GameFee     float32 `gorm:"-" json:"game_fee"`
	// JackpotFree float32 `gorm:"-" json:"jackpot_fee"
	JpChips int64 `gorm:"-" json:"jp_chips"`
}

func (Game) TableName() string {
	return "games"
}

type Layout struct {
	Col     int32 `json:"col"`
	Row     int32 `json:"row"`
	ColSpan int32 `json:"colSpan"`
	RowSpan int32 `json:"rowSpan"`
}

const (
	BucketAvatar   = "avatar"
	BucketBanners  = "banners"
	AvatarFileName = "%s_image"
	LinkFanpageFB  = "https://www.facebook.com/"
	LinkGroupFB    = "https://www.facebook.com/"
)

func InterfaceToString(inf interface{}) string {
	if inf == nil {
		return ""
	}
	str, ok := inf.(string)
	if !ok {
		return ""
	}
	return str
}

func ToInt64(inf interface{}, def int64) int64 {
	if inf == nil {
		return def
	}
	switch v := inf.(type) {
	case int:
		return int64(inf.(int))
	case int64:
		return inf.(int64)
	case string:
		str := inf.(string)
		i, _ := strconv.ParseInt(str, 10, 64)
		return i
	case float64:
		return int64(inf.(float64))
	default:
		fmt.Printf("I don't know about type %T!\n", v)
	}
	return def
}

type WalletAction string

const (
	WalletActionBankTopup   WalletAction = "bank_topup"
	WalletActionDailyReward WalletAction = "daily_reward"
	WalletActionFreeChip    WalletAction = "free_chip"
	WalletActionGiftCode    WalletAction = "gift_code"
	WalletActionIAPTopUp    WalletAction = "iap_topup"
	WalletActionReferReward WalletAction = "refer_reward"
	WalletActionUserGift    WalletAction = "user_gift"
)

func (w WalletAction) String() string {
	return string(w)
}

var MapWalletAction map[string]bool
