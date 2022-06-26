package entity

import (
	"fmt"
	"strconv"

	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
)

const (
	ModuleName         = "lobby"
	AutoPrefix         = "CGPD"
	AutoPrefixFacebook = "CGPF"
)

const (
	MIN_LENGTH_PASSWORD = 6
)

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
			Layout: &pb.Layout{
				Col:     g.Layout.Col,
				Row:     g.Layout.Row,
				ColSpan: g.Layout.ColSpan,
				RowSpan: g.Layout.RowSpan,
			},
			LobbyId: g.LobbyId,
		})
	}

	return pbGames
}

type Game struct {
	Code    string `json:"code"`
	Layout  Layout `json:"layout"`
	LobbyId string `json:"lobbyId"`
	MinChip int    `json:"minChip"`
	Enable  bool   `json:"enable"`
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
