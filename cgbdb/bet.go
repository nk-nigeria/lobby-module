package cgbdb

import (
	"context"
	"database/sql"
	"errors"

	"github.com/nakamaFramework/cgb-lobby-module/entity"
)

func AddBet(ctx context.Context, db *sql.DB, bet *entity.Bet) error {
	gDB, err := NewGormContext(ctx, db)
	if err != nil {
		return err
	}
	err = gDB.Model(bet).Create(bet).Error
	return err
}

func UpdateBet(ctx context.Context, db *sql.DB, bet *entity.Bet) error {
	if bet.Id <= 0 {
		return errors.New("missing id")
	}
	gDB, err := NewGormContext(ctx, db)
	if err != nil {
		return err
	}
	err = gDB.Model(bet).Updates(bet).Error
	return err
}

func ReadBet(ctx context.Context, db *sql.DB, id int64) (*entity.Bet, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	gDB, err := NewGormContext(ctx, db)
	if err != nil {
		return nil, err
	}
	bet := &entity.Bet{}
	err = gDB.Model(bet).First(bet, id).Error
	fillAgbet(bet)
	return bet, err
}

func QueryBet(ctx context.Context, db *sql.DB, limit, offset int64, query interface{}, args ...interface{}) ([]entity.Bet, int64, error) {
	gDB, err := NewGormContext(ctx, db)
	if err != nil {
		return nil, 0, err
	}
	ml := make([]entity.Bet, 0)
	tx := gDB.Model(new(entity.Bet)).Where(query, args...)
	total := int64(0)
	err = tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return ml, total, nil
	}
	err = tx.Order("id desc").Order("game_id desc").Find(&ml).Error
	for idx, v := range ml {
		fillAgbet(&v)
		ml[idx] = v
	}
	return ml, 0, err
}

func DeleteBet(ctx context.Context, db *sql.DB, id int64) (*entity.Bet, error) {
	if id <= 0 {
		return nil, errors.New("missing id")
	}
	gDB, err := NewGormContext(ctx, db)
	if err != nil {
		return nil, err
	}
	betDeleted, err := ReadBet(ctx, db, id)
	if err != nil {
		return nil, err
	}
	err = gDB.Delete(entity.Bet{}, id).Error
	if err != nil {
		return nil, err
	}
	return betDeleted, nil
}

func fillAgbet(bet *entity.Bet) *entity.Bet {
	if bet == nil {
		return nil
	}
	bet.AGJoin = int(float64(bet.MarkUnit) * float64(bet.Xjoin))
	bet.AGPlaynow = int(float64(bet.MarkUnit) * float64(bet.Xplaynow))
	bet.AGLeave = int(float64(bet.MarkUnit) * float64(bet.Xleave))
	bet.AGFee = int(float64(bet.MarkUnit) * float64(bet.Xfee))
	return bet
}
