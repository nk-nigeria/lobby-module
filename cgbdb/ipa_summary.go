package cgbdb

import (
	"database/sql"
	"time"

	"github.com/nakamaFramework/cgb-lobby-module/entity"
	"gorm.io/gorm"
)

type IAPSummary struct {
	ID         uint      `gorm:"primarykey" json:"id,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
	UserId     string    `gorm:"uniqueIndex"  json:"user_id,omitempty"`
	TotalTopup int64     `json:"total_topup,omitempty"`
	VipPoint   int64     `json:"vip_point,omitempty"`
}

func UpdateTopupSummary(db *sql.DB, userId string, chips int64) error {
	iapSummary := IAPSummary{
		UserId: userId,
	}
	gDb, err := NewGorm(db)
	if err != nil {
		return err
	}
	vipPoint := entity.ExchangeChipsToVipPoint(chips)
	tx := gDb.Model(&iapSummary).Where("user_id=?", userId).
		Updates(map[string]interface{}{
			"total_topup": gorm.Expr("total_topup + ?", chips),
			"vip_point":   gorm.Expr("vip_point + ?", vipPoint),
		})
	if tx.Error == nil && tx.RowsAffected > 0 {
		return nil
	}
	iapSummary = IAPSummary{
		UserId:     userId,
		TotalTopup: chips,
		VipPoint:   vipPoint,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	tx = gDb.Model(&iapSummary).Create(&iapSummary)
	return tx.Error
}
