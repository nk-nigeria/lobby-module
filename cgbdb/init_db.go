package cgbdb

import (
	"context"
	"database/sql"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var mapDb = make(map[*sql.DB]*gorm.DB)

func AutoMigrate(db *sql.DB) {
	gDb, err := NewGorm(db)
	if err != nil {
		return
	}
	gDb.AutoMigrate(new(IAPSummary))
}

func NewGorm(db *sql.DB) (*gorm.DB, error) {
	gormDb, found := mapDb[db]
	if found {
		return gormDb, nil
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	mapDb[db] = gormDB
	return gormDB, err
}

func NewGormContext(ctx context.Context, db *sql.DB) (*gorm.DB, error) {
	gormDb, found := mapDb[db]
	if found {
		return gormDb, nil
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	mapDb[db] = gormDB
	return gormDB.WithContext(ctx), err
}
