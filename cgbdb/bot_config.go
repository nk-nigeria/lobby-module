package cgbdb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

type BotJoinRule struct {
	ID            int    `json:"id"`
	GameCode      string `json:"game_code"`
	MinBet        int64  `json:"min_bet"`
	MaxBet        int64  `json:"max_bet"`
	MinUsers      int    `json:"min_users"`
	MaxUsers      int    `json:"max_users"`
	RandomTimeMin int    `json:"random_time_min"`
	RandomTimeMax int    `json:"random_time_max"`
	JoinPercent   int    `json:"join_percent"`
	IsActive      bool   `json:"is_active"`
}

type BotLeaveRule struct {
	ID           int    `json:"id"`
	GameCode     string `json:"game_code"`
	MinBet       int64  `json:"min_bet"`
	MaxBet       int64  `json:"max_bet"`
	LastResult   int    `json:"last_result"`
	LeavePercent int    `json:"leave_percent"`
	IsActive     bool   `json:"is_active"`
}

type BotCreateTableRule struct {
	ID              int    `json:"id"`
	GameCode        string `json:"game_code"`
	MinBet          int64  `json:"min_bet"`
	MaxBet          int64  `json:"max_bet"`
	MinActiveTables int    `json:"min_active_tables"`
	MaxActiveTables int    `json:"max_active_tables"`
	WaitTimeMin     int    `json:"wait_time_min"`
	WaitTimeMax     int    `json:"wait_time_max"`
	RetryWaitMin    int    `json:"retry_wait_min"`
	RetryWaitMax    int    `json:"retry_wait_max"`
	IsActive        bool   `json:"is_active"`
}

type BotGroupRule struct {
	ID       int    `json:"id"`
	GameCode string `json:"game_code"`
	VIPMin   int    `json:"vip_min"`
	VIPMax   int    `json:"vip_max"`
	MCBMin   int64  `json:"mcb_min"`
	MCBMax   int64  `json:"mcb_max"`
	IsActive bool   `json:"is_active"`
}

type BotConfig struct {
	GameCode            string               `json:"game_code"`
	BotJoinRules        []BotJoinRule        `json:"bot_join_rules"`
	BotLeaveRules       []BotLeaveRule       `json:"bot_leave_rules"`
	BotCreateTableRules []BotCreateTableRule `json:"bot_create_table_rules"`
	BotGroupRules       []BotGroupRule       `json:"bot_group_rules"`
}

// GetBotJoinRules retrieves bot join rules for a specific game
func GetBotJoinRules(ctx context.Context, logger runtime.Logger, db *sql.DB, gameCode string) ([]BotJoinRule, error) {
	query := `
		SELECT id, game_code, min_bet, max_bet, min_users, max_users, 
		       random_time_min, random_time_max, join_percent, is_active
		FROM bot_join_rules 
		WHERE game_code = $1 AND is_active = true
		ORDER BY min_bet ASC
	`

	rows, err := db.QueryContext(ctx, query, gameCode)
	if err != nil {
		logger.Error("Failed to query bot join rules: %v", err)
		return nil, err
	}
	defer rows.Close()

	var rules []BotJoinRule
	for rows.Next() {
		var rule BotJoinRule
		err := rows.Scan(
			&rule.ID, &rule.GameCode, &rule.MinBet, &rule.MaxBet,
			&rule.MinUsers, &rule.MaxUsers, &rule.RandomTimeMin,
			&rule.RandomTimeMax, &rule.JoinPercent, &rule.IsActive,
		)
		if err != nil {
			logger.Error("Failed to scan bot join rule: %v", err)
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// SaveBotJoinRule saves or updates a bot join rule
func SaveBotJoinRule(ctx context.Context, logger runtime.Logger, db *sql.DB, rule *BotJoinRule) error {
	if rule.ID == 0 {
		// Insert new rule
		query := `
			INSERT INTO bot_join_rules (game_code, min_bet, max_bet, min_users, max_users, 
			                           random_time_min, random_time_max, join_percent, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`
		err := db.QueryRowContext(ctx, query,
			rule.GameCode, rule.MinBet, rule.MaxBet, rule.MinUsers, rule.MaxUsers,
			rule.RandomTimeMin, rule.RandomTimeMax, rule.JoinPercent, rule.IsActive,
		).Scan(&rule.ID)
		if err != nil {
			logger.Error("Failed to insert bot join rule: %v", err)
			return err
		}
	} else {
		// Update existing rule
		query := `
			UPDATE bot_join_rules 
			SET min_bet = $2, max_bet = $3, min_users = $4, max_users = $5,
			    random_time_min = $6, random_time_max = $7, join_percent = $8, is_active = $9
			WHERE id = $1
		`
		_, err := db.ExecContext(ctx, query,
			rule.ID, rule.MinBet, rule.MaxBet, rule.MinUsers, rule.MaxUsers,
			rule.RandomTimeMin, rule.RandomTimeMax, rule.JoinPercent, rule.IsActive,
		)
		if err != nil {
			logger.Error("Failed to update bot join rule: %v", err)
			return err
		}
	}
	return nil
}

// GetBotLeaveRules retrieves bot leave rules for a specific game
func GetBotLeaveRules(ctx context.Context, logger runtime.Logger, db *sql.DB, gameCode string) ([]BotLeaveRule, error) {
	query := `
		SELECT id, game_code, min_bet, max_bet, last_result, leave_percent, is_active
		FROM bot_leave_rules 
		WHERE game_code = $1 AND is_active = true
		ORDER BY min_bet ASC
	`

	rows, err := db.QueryContext(ctx, query, gameCode)
	if err != nil {
		logger.Error("Failed to query bot leave rules: %v", err)
		return nil, err
	}
	defer rows.Close()

	var rules []BotLeaveRule
	for rows.Next() {
		var rule BotLeaveRule
		err := rows.Scan(
			&rule.ID, &rule.GameCode, &rule.MinBet, &rule.MaxBet,
			&rule.LastResult, &rule.LeavePercent, &rule.IsActive,
		)
		if err != nil {
			logger.Error("Failed to scan bot leave rule: %v", err)
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// SaveBotLeaveRule saves or updates a bot leave rule
func SaveBotLeaveRule(ctx context.Context, logger runtime.Logger, db *sql.DB, rule *BotLeaveRule) error {
	if rule.ID == 0 {
		// Insert new rule
		query := `
			INSERT INTO bot_leave_rules (game_code, min_bet, max_bet, last_result, leave_percent, is_active)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`
		err := db.QueryRowContext(ctx, query,
			rule.GameCode, rule.MinBet, rule.MaxBet, rule.LastResult, rule.LeavePercent, rule.IsActive,
		).Scan(&rule.ID)
		if err != nil {
			logger.Error("Failed to insert bot leave rule: %v", err)
			return err
		}
	} else {
		// Update existing rule
		query := `
			UPDATE bot_leave_rules 
			SET min_bet = $2, max_bet = $3, last_result = $4, leave_percent = $5, is_active = $6
			WHERE id = $1
		`
		_, err := db.ExecContext(ctx, query,
			rule.ID, rule.MinBet, rule.MaxBet, rule.LastResult, rule.LeavePercent, rule.IsActive,
		)
		if err != nil {
			logger.Error("Failed to update bot leave rule: %v", err)
			return err
		}
	}
	return nil
}

// GetBotCreateTableRules retrieves bot create table rules for a specific game
func GetBotCreateTableRules(ctx context.Context, logger runtime.Logger, db *sql.DB, gameCode string) ([]BotCreateTableRule, error) {
	query := `
		SELECT id, game_code, min_bet, max_bet, min_active_tables, max_active_tables,
		       wait_time_min, wait_time_max, retry_wait_min, retry_wait_max, is_active
		FROM bot_create_table_rules 
		WHERE game_code = $1 AND is_active = true
		ORDER BY min_bet ASC
	`

	rows, err := db.QueryContext(ctx, query, gameCode)
	if err != nil {
		logger.Error("Failed to query bot create table rules: %v", err)
		return nil, err
	}
	defer rows.Close()

	var rules []BotCreateTableRule
	for rows.Next() {
		var rule BotCreateTableRule
		err := rows.Scan(
			&rule.ID, &rule.GameCode, &rule.MinBet, &rule.MaxBet,
			&rule.MinActiveTables, &rule.MaxActiveTables, &rule.WaitTimeMin,
			&rule.WaitTimeMax, &rule.RetryWaitMin, &rule.RetryWaitMax, &rule.IsActive,
		)
		if err != nil {
			logger.Error("Failed to scan bot create table rule: %v", err)
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// SaveBotCreateTableRule saves or updates a bot create table rule
func SaveBotCreateTableRule(ctx context.Context, logger runtime.Logger, db *sql.DB, rule *BotCreateTableRule) error {
	if rule.ID == 0 {
		// Insert new rule
		query := `
			INSERT INTO bot_create_table_rules (game_code, min_bet, max_bet, min_active_tables, max_active_tables,
			                                  wait_time_min, wait_time_max, retry_wait_min, retry_wait_max, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id
		`
		err := db.QueryRowContext(ctx, query,
			rule.GameCode, rule.MinBet, rule.MaxBet, rule.MinActiveTables, rule.MaxActiveTables,
			rule.WaitTimeMin, rule.WaitTimeMax, rule.RetryWaitMin, rule.RetryWaitMax, rule.IsActive,
		).Scan(&rule.ID)
		if err != nil {
			logger.Error("Failed to insert bot create table rule: %v", err)
			return err
		}
	} else {
		// Update existing rule
		query := `
			UPDATE bot_create_table_rules 
			SET min_bet = $2, max_bet = $3, min_active_tables = $4, max_active_tables = $5,
			    wait_time_min = $6, wait_time_max = $7, retry_wait_min = $8, retry_wait_max = $9, is_active = $10
			WHERE id = $1
		`
		_, err := db.ExecContext(ctx, query,
			rule.ID, rule.MinBet, rule.MaxBet, rule.MinActiveTables, rule.MaxActiveTables,
			rule.WaitTimeMin, rule.WaitTimeMax, rule.RetryWaitMin, rule.RetryWaitMax, rule.IsActive,
		)
		if err != nil {
			logger.Error("Failed to update bot create table rule: %v", err)
			return err
		}
	}
	return nil
}

// GetBotGroupRules retrieves bot group rules for a specific game
func GetBotGroupRules(ctx context.Context, logger runtime.Logger, db *sql.DB, gameCode string) ([]BotGroupRule, error) {
	query := `
		SELECT id, game_code, vip_min, vip_max, mcb_min, mcb_max, is_active
		FROM bot_group_rules 
		WHERE game_code = $1 AND is_active = true
		ORDER BY vip_min ASC
	`

	rows, err := db.QueryContext(ctx, query, gameCode)
	if err != nil {
		logger.Error("Failed to query bot group rules: %v", err)
		return nil, err
	}
	defer rows.Close()

	var rules []BotGroupRule
	for rows.Next() {
		var rule BotGroupRule
		err := rows.Scan(
			&rule.ID, &rule.GameCode, &rule.VIPMin, &rule.VIPMax, &rule.MCBMin, &rule.MCBMax, &rule.IsActive,
		)
		if err != nil {
			logger.Error("Failed to scan bot group rule: %v", err)
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// SaveBotGroupRule saves or updates a bot group rule
func SaveBotGroupRule(ctx context.Context, logger runtime.Logger, db *sql.DB, rule *BotGroupRule) error {
	if rule.ID == 0 {
		// Insert new rule
		query := `
			INSERT INTO bot_group_rules (game_code, vip_min, vip_max, mcb_min, mcb_max, is_active)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`
		err := db.QueryRowContext(ctx, query,
			rule.GameCode, rule.VIPMin, rule.VIPMax, rule.MCBMin, rule.MCBMax, rule.IsActive,
		).Scan(&rule.ID)
		if err != nil {
			logger.Error("Failed to insert bot group rule: %v", err)
			return err
		}
	} else {
		// Update existing rule
		query := `
			UPDATE bot_group_rules 
			SET vip_min = $2, vip_max = $3, mcb_min = $4, mcb_max = $5, is_active = $6
			WHERE id = $1
		`
		_, err := db.ExecContext(ctx, query,
			rule.ID, rule.VIPMin, rule.VIPMax, rule.MCBMin, rule.MCBMax, rule.IsActive,
		)
		if err != nil {
			logger.Error("Failed to update bot group rule: %v", err)
			return err
		}
	}
	return nil
}

// DeleteBotRule deactivates a bot rule by setting is_active = false
func DeleteBotRule(ctx context.Context, logger runtime.Logger, db *sql.DB, tableName string, ruleID int) error {
	query := fmt.Sprintf("UPDATE %s SET is_active = false WHERE id = $1", tableName)
	_, err := db.ExecContext(ctx, query, ruleID)
	if err != nil {
		logger.Error("Failed to delete bot rule: %v", err)
		return err
	}
	return nil
}
