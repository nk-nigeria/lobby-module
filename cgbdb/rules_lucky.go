package cgbdb

import (
	"context"
	"database/sql"
	"time"

	pb "github.com/nakamaFramework/cgp-common/proto"
)

type rulesLucky struct {
	Id              uint      `gorm:"primarykey" json:"id,omitempty"`
	CreateAt        time.Time `json:"create_at,omitempty"`
	GameCode        string    `json:"game_code,omitempty"`
	EmitEventAtUnix int64     `json:"emit_event_at_unix,omitempty"`
	DeletedAt       int64     `json:"deleted_at,omitempty"`
	RtpMin          int64     `json:"rtp_min,omitempty"`
	RtpMax          int64     `json:"rtp_max,omitempty"`
	MarkMin         int64     `json:"mark_min,omitempty"`
	MarkMax         int64     `json:"mark_max,omitempty"`
	VipMin          int64     `json:"vip_min,omitempty"`
	VipMax          int64     `json:"vip_max,omitempty"`
	WinMarkRatioMin int64     `json:"win_mark_ratio_min,omitempty"`
	WinMarkRatioMax int64     `json:"win_mark_ratio_max,omitempty"`
	ReDeal          int64     `json:"re_deal,omitempty"`
}

func (*rulesLucky) TableName() string {
	return "rules_lucky"
}

func (r *rulesLucky) Copy(rule *pb.RuleLucky) {
	r.Id = uint(rule.Id)
	r.GameCode = rule.GameCode
	r.EmitEventAtUnix = rule.EmitEventAtUnix
	r.DeletedAt = rule.DeletedAt
	r.RtpMax = rule.Rtp.Max
	r.RtpMin = rule.Rtp.Min
	r.MarkMin = rule.Mark.Min
	r.RtpMax = rule.Mark.Max
	r.VipMin = rule.Vip.Min
	r.VipMax = rule.Vip.Max
	r.WinMarkRatioMin = rule.WinMarkRatio.Min
	r.WinMarkRatioMax = rule.WinMarkRatio.Max
	r.ReDeal = rule.ReDeal
}

func (r *rulesLucky) Trasnfer(rule *pb.RuleLucky) {
	rule.Id = int64(r.Id)
	rule.GameCode = r.GameCode
	rule.EmitEventAtUnix = r.EmitEventAtUnix
	rule.DeletedAt = r.DeletedAt
	rule.Rtp = &pb.Range{Min: r.RtpMin, Max: r.RtpMax}
	rule.Mark = &pb.Range{Min: r.MarkMin, Max: r.MarkMax}
	rule.Vip = &pb.Range{Min: r.VipMin, Max: r.VipMax}
	rule.WinMarkRatio = &pb.Range{Min: r.WinMarkRatioMin, Max: r.WinMarkRatioMax}
	rule.ReDeal = r.ReDeal
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
			"rtp_min":            rule.Rtp.Min,
			"rtp_max":            rule.Rtp.Max,
			"vip_min":            rule.Vip.Min,
			"vip_max":            rule.Vip.Max,
			"mark_min":           rule.Mark.Min,
			"mark_max":           rule.Mark.Max,
			"win_mark_ratio_min": rule.WinMarkRatio.Min,
			"win_mark_ratio_max": rule.WinMarkRatio.Max,
			"re_deal":            rule.ReDeal,
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
