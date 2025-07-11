package cgbdb

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/nk-nigeria/lobby-module/entity"
	"gorm.io/gorm/clause"
)

func AddGame(ctx context.Context, db *sql.DB, game *entity.Game) error {
	gDB, err := NewGormContext(ctx, db)
	if err != nil {
		return err
	}
	// game.ID = 0
	// return gDB.Model(game).Create(game).Error
	return gDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).Create(game).Error
}

func ListGames(ctx context.Context, db *sql.DB) ([]entity.Game, error) {
	gDB, err := NewGormContext(ctx, db)
	if err != nil {
		return nil, err
	}
	ml := make([]entity.Game, 0)
	err = gDB.Model(new(entity.Game)).Find(&ml).Error
	for idx, game := range ml {
		game.LobbyId = strconv.FormatInt(int64(game.ID), 10)
		ml[idx] = game
	}
	return ml, err
}
