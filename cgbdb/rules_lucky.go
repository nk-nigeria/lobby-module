package cgbdb

import (
	"context"
	"database/sql"
	"time"

	pb "github.com/ciaolink-game-platform/cgp-common/proto"
)

type rulesLucky struct {
	Id         uint      `gorm:"primarykey" json:"id,omitempty"`
	CreateAt   time.Time `json:"create_at,omitempty"`
	GameCode   string    `json:"game_code,omitempty"`
	CoRateMin  float32   `json:"co_rate_min,omitempty"`
	CoRateMax  float32   `json:"co_rate_max,omitempty"`
	CiMin      float32   `json:"ci_min,omitempty"`
	CiMax      float32   `json:"ci_max,omitempty"`
	CoIndayMin float32   `json:"co_inday_min,omitempty"`
	CoIndayMax float32   `json:"co_inday_max,omitempty"`
	Base_1     int       `json:"base_1,omitempty"`
	Base_2     int       `json:"base_2,omitempty"`
	Base_3     int       `json:"base_3,omitempty"`
	Base_4     int       `json:"base_4,omitempty"`
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
}

func InsertRulesLucky(ctx context.Context, db *sql.DB, rule *pb.RuleLucky) error {
	r := &rulesLucky{}
	r.Copy(rule)
	gOrm, err := NewGorm(db)
	if err != nil {
		return err
	}
	r.CreateAt = time.Now()
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
	tx := gOrm.Model(new(rulesLucky)).Where("id = ?", rule.Id).
		Updates(map[string]interface{}{
			"co_rate_min":  rule.CoRateMin,
			"co_rate_max":  rule.CoRateMax,
			"ci_min":       rule.CiMin,
			"ci_max":       rule.CiMax,
			"co_inday_min": rule.CoIndayMin,
			"co_inday_max": rule.CoIndayMax,
			"base_1":       rule.Base_1,
			"base_2":       rule.Base_2,
			"base_3":       rule.Base_3,
			"base_4":       rule.Base_4,
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
	if rule.Id > 0 {
		tx = tx.Where("id = ?", rule.Id)
	}
	tx = tx.Order("id ASC").Find(&ml)
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
	tx := gOrm.Delete(&rulesLucky{}, id)
	return tx.Error
}
