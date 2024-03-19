package cgbdb

import (
	"context"
	"database/sql"
	"time"

	pb "github.com/ciaolink-game-platform/cgp-common/proto"
)

type rulesLucky struct {
	Id              uint      `gorm:"primarykey" json:"id,omitempty"`
	CreateAt        time.Time `json:"create_at,omitempty"`
	GameCode        string    `json:"game_code,omitempty"`
	CoRateMin       float32   `json:"co_rate_min,omitempty"`
	CoRateMax       float32   `json:"co_rate_max,omitempty"`
	CiMin           float32   `json:"ci_min,omitempty"`
	CiMax           float32   `json:"ci_max,omitempty"`
	CoIndayMin      float32   `json:"co_inday_min,omitempty"`
	CoIndayMax      float32   `json:"co_inday_max,omitempty"`
	Base_1          int       `json:"base_1,omitempty"`
	Base_2          int       `json:"base_2,omitempty"`
	Base_3          int       `json:"base_3,omitempty"`
	Base_4          int       `json:"base_4,omitempty"`
	EmitEventAtUnix int64     `json:"emit_event_at_unix,omitempty"`
	DeletedAt       int64     `json:"deleted_at,omitempty"`
}

func (*rulesLucky) TableName() string {
	return "rules_lucky"
}

func (r *rulesLucky) Copy(rule *pb.RuleLucky) {
	r.Id = uint(rule.Id)
	r.GameCode = rule.GameCode
	r.CoRateMin = rule.CoRateMin
	r.CoRateMax = rule.CoRateMax
	r.CiMin = rule.CiMin
	r.CiMax = rule.CiMax
	r.CoIndayMin = rule.CoIndayMin
	r.CoIndayMax = rule.CoIndayMax
	r.Base_1 = int(rule.Base_1)
	r.Base_2 = int(rule.Base_2)
	r.Base_3 = int(rule.Base_3)
	r.Base_4 = int(rule.Base_4)
	r.EmitEventAtUnix = rule.EmitEventAtUnix
	r.DeletedAt = rule.DeletedAt
}

func (r *rulesLucky) Trasnfer(rule *pb.RuleLucky) {
	rule.Id = int64(r.Id)
	rule.GameCode = r.GameCode
	rule.CoRateMin = r.CoRateMin
	rule.CoRateMax = r.CoRateMax
	rule.CiMin = r.CiMin
	rule.CiMax = r.CiMax
	rule.CoIndayMin = r.CoIndayMin
	rule.CoIndayMax = r.CoIndayMax
	rule.Base_1 = int64(r.Base_1)
	rule.Base_2 = int64(r.Base_2)
	rule.Base_3 = int64(r.Base_3)
	rule.Base_4 = int64(r.Base_4)
	rule.EmitEventAtUnix = r.EmitEventAtUnix
	rule.DeletedAt = r.DeletedAt
}

func InsertRulesLucky(ctx context.Context, db *sql.DB, rule *pb.RuleLucky) error {
	r := &rulesLucky{}
	r.Copy(rule)
	gOrm, err := NewGorm(db)
	if err != nil {
		return err
	}
	r.CreateAt = time.Now()
	r.DeletedAt = 0
	r.EmitEventAtUnix = 1
	return gOrm.Create(r).Error
}

func UpdateRulesLucky(ctx context.Context, db *sql.DB, rule *pb.RuleLucky) (*pb.RuleLucky, error) {
	gOrm, err := NewGorm(db)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return &pb.RuleLucky{}, nil
	}
	tx := gOrm.Model(new(rulesLucky)).Where("id = ? and deleted_at = 0", rule.Id).
		Updates(map[string]interface{}{
			"co_rate_min":        rule.CoRateMin,
			"co_rate_max":        rule.CoRateMax,
			"ci_min":             rule.CiMin,
			"ci_max":             rule.CiMax,
			"co_inday_min":       rule.CoIndayMin,
			"co_inday_max":       rule.CoIndayMax,
			"base_1":             rule.Base_1,
			"base_2":             rule.Base_2,
			"base_3":             rule.Base_3,
			"base_4":             rule.Base_4,
			"emit_event_at_unix": 1,
		})
	return rule, tx.Error
}

func QueryRulesLucky(ctx context.Context, db *sql.DB, rule *pb.RuleLucky) ([]*pb.RuleLucky, error) {
	gOrm, err := NewGorm(db)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		rule = &pb.RuleLucky{}
	}
	ml := make([]rulesLucky, 0)
	tx := gOrm.Model(new(rulesLucky))
	args := make([]any, 0)
	query := "1=1"
	if rule.DeletedAt == 0 {
		query += " AND deleted_at = ?"
		args = append(args, rule.DeletedAt)
	}
	if rule.Id > 0 {
		query += " AND id = ?"
		args = append(args, rule.Id)
	}
	if len(rule.GameCode) > 0 {
		query += " AND game_code = ?"
		args = append(args, rule.GameCode)
	}
	if rule.EmitEventAtUnix > 0 {
		query += " AND emit_event_at_unix = ?"
		args = append(args, rule.EmitEventAtUnix)
	}
	tx = tx.Where(query, args...).Order("id ASC").Find(&ml)
	if tx.Error != nil {
		return nil, tx.Error
	}
	list := make([]*pb.RuleLucky, 0, len(ml))
	for _, r := range ml {
		v := &pb.RuleLucky{}
		r.Trasnfer(v)
		list = append(list, v)
	}
	return list, nil
}

func DeleteRulesLucky(ctx context.Context, db *sql.DB, id int64) error {
	gOrm, err := NewGorm(db)
	if err != nil {
		return err
	}
	if id <= 0 {
		return nil
	}
	tx := gOrm.Model(new(rulesLucky)).Where("id=?", id).Updates(map[string]interface{}{"deleted_at": time.Now().Unix(), "emit_event_at_unix": 1})
	return tx.Error
}

func UpdateEmitEventLucky(ctx context.Context, db *sql.DB, rule *pb.RuleLucky) error {
	if len(rule.GameCode) == 0 {
		return nil
	}
	gOrm, err := NewGorm(db)
	if err != nil {
		return err
	}
	tx := gOrm.Model(new(rulesLucky)).Where("game_code = ?", rule.GameCode).
		Updates(map[string]interface{}{"emit_event_at_unix": time.Now().Unix()})
	return tx.Error

}
