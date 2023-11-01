package cgbdb

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type IAPSummary struct {
	ID         uint      `gorm:"primarykey" json:"id,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
	UserId     string    `json:"user_id,omitempty"`
	TotalTopup int64     `gorm:"uniqueIndex" json:"total_topup,omitempty"`
}

func UpdateTopupSummary(db *sql.DB, userId string, chips int64) error {
	iapSummary := IAPSummary{
		UserId: userId,
		// TotalTopup: chips,
		// CreatedAt:  time.Now(),
		// UpdatedAt:  time.Now(),
	}
	gDb, err := NewGorm(db)
	if err != nil {
		return err
	}
	tx := gDb.Model(&iapSummary).Where("user_id=?", userId).First(&iapSummary)
	if tx.Error == gorm.ErrRecordNotFound {
		iapSummary := IAPSummary{
			UserId:     userId,
			TotalTopup: chips,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		tx = gDb.Model(&iapSummary).Create(&iapSummary)
		return tx.Error
	}
	tx = gDb.Model(&iapSummary).Where("user_id=?", userId).
		Update("total_topup", gorm.Expr("total_topup + ?", chips))
	return tx.Error
}
