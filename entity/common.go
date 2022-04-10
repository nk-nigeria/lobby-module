package entity

const (
	ModuleName          = "lobby"
	MIN_LENGTH_PASSWORD = 6
)

type CustomUser struct {
	Id       string
	UserId   string
	UserName string
}
