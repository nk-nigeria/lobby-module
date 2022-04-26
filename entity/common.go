package entity

import (
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
)

const (
	ModuleName          = "lobby"
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
