package entity

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

type Bet struct {
	Amount  int  `json:"amount"`
	MinChip int  `json:"minChip,omitempty"`
	Enable  bool `json:"enable"`
}

type Bets struct {
	List []Bet `json:"bets"`
}
