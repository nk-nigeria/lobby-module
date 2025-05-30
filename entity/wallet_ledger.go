package entity

import "time"

type WalletLedgerListCursor struct {
	UserId         string
	CreateTime     time.Time
	Id             string
	IsNext         bool
	MetaAction     []string
	MetaBankAction []string
}

type WalletLedger struct {
	ID         string                 `json:"id"`
	UserId     string                 `json:"userId"`
	CreateTime int64                  `json:"createdTime"`
	UpdateTime int64                  `json:"updateTime"`
	Changeset  map[string]int64       `json:"changeset"`
	Metadata   map[string]interface{} `json:"metadata"`
}

func (w *WalletLedger) GetID() string {
	return w.ID
}
func (w *WalletLedger) GetUserID() string {
	return w.UserId
}
func (w *WalletLedger) GetCreateTime() int64 {
	return w.CreateTime
}
func (w *WalletLedger) GetUpdateTime() int64 {
	return w.UpdateTime
}
func (w *WalletLedger) GetChangeset() map[string]int64 {
	return w.Changeset
}
func (w *WalletLedger) GetMetadata() map[string]interface{} {
	return w.Metadata
}
