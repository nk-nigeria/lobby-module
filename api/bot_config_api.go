package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nk-nigeria/lobby-module/cgbdb"
	"google.golang.org/protobuf/proto"
)

// RpcGetBotConfig lấy cấu hình bot hiện tại
func RpcGetBotConfig(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		// Parse request payload
		var request map[string]interface{}
		if err := json.Unmarshal([]byte(payload), &request); err != nil {
			logger.Error("Failed to unmarshal request: %v", err)
			return "", err
		}

		gameCode, ok := request["game_code"].(string)
		if !ok || gameCode == "" {
			logger.Error("Invalid or missing game_code in request")
			return "", runtime.NewError("Invalid game_code", 3)
		}

		// Get bot config from database
		config, err := getBotConfigFromDB(ctx, logger, db, gameCode)
		if err != nil {
			logger.Error("Failed to get bot config: %v", err)
			return "", err
		}

		responseBytes, err := json.Marshal(config)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return "", err
		}

		return string(responseBytes), nil
	}
}

// RpcUpdateBotConfig cập nhật cấu hình bot
func RpcUpdateBotConfig(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		// Parse request payload
		var config cgbdb.BotConfig
		if err := json.Unmarshal([]byte(payload), &config); err != nil {
			logger.Error("Failed to unmarshal bot config request: %v", err)
			return "", err
		}

		// Validate config
		if err := validateBotConfig(&config); err != nil {
			logger.Error("Invalid bot config: %v", err)
			return "", runtime.NewError("Invalid bot config: "+err.Error(), 3)
		}

		// Save to database
		if err := saveBotConfigToDB(ctx, logger, db, &config); err != nil {
			logger.Error("Failed to save bot config to database: %v", err)
			return "", err
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Bot config updated successfully",
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return "", err
		}

		return string(responseBytes), nil
	}
}

// getBotConfigFromDB retrieves bot configuration from database
func getBotConfigFromDB(ctx context.Context, logger runtime.Logger, db *sql.DB, gameCode string) (*cgbdb.BotConfig, error) {
	config := &cgbdb.BotConfig{
		GameCode:            gameCode,
		BotJoinRules:        []cgbdb.BotJoinRule{},
		BotLeaveRules:       []cgbdb.BotLeaveRule{},
		BotCreateTableRules: []cgbdb.BotCreateTableRule{},
		BotGroupRules:       []cgbdb.BotGroupRule{},
	}

	// Get bot join rules
	joinRules, err := cgbdb.GetBotJoinRules(ctx, logger, db, gameCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot join rules: %v", err)
	}
	config.BotJoinRules = joinRules

	// Get bot leave rules
	leaveRules, err := cgbdb.GetBotLeaveRules(ctx, logger, db, gameCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot leave rules: %v", err)
	}
	config.BotLeaveRules = leaveRules

	// Get bot create table rules
	createTableRules, err := cgbdb.GetBotCreateTableRules(ctx, logger, db, gameCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot create table rules: %v", err)
	}
	config.BotCreateTableRules = createTableRules

	// Get bot group rules
	groupRules, err := cgbdb.GetBotGroupRules(ctx, logger, db, gameCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot group rules: %v", err)
	}
	config.BotGroupRules = groupRules

	return config, nil
}

// saveBotConfigToDB saves bot configuration to database
func saveBotConfigToDB(ctx context.Context, logger runtime.Logger, db *sql.DB, config *cgbdb.BotConfig) error {
	// Save bot join rules
	for _, rule := range config.BotJoinRules {
		rule.GameCode = config.GameCode
		if err := cgbdb.SaveBotJoinRule(ctx, logger, db, &rule); err != nil {
			return fmt.Errorf("failed to save bot join rule: %v", err)
		}
	}

	// Save bot leave rules
	for _, rule := range config.BotLeaveRules {
		rule.GameCode = config.GameCode
		if err := cgbdb.SaveBotLeaveRule(ctx, logger, db, &rule); err != nil {
			return fmt.Errorf("failed to save bot leave rule: %v", err)
		}
	}

	// Save bot create table rules
	for _, rule := range config.BotCreateTableRules {
		rule.GameCode = config.GameCode
		if err := cgbdb.SaveBotCreateTableRule(ctx, logger, db, &rule); err != nil {
			return fmt.Errorf("failed to save bot create table rule: %v", err)
		}
	}

	// Save bot group rules
	for _, rule := range config.BotGroupRules {
		rule.GameCode = config.GameCode
		if err := cgbdb.SaveBotGroupRule(ctx, logger, db, &rule); err != nil {
			return fmt.Errorf("failed to save bot group rule: %v", err)
		}
	}

	return nil
}

// validateBotConfig kiểm tra tính hợp lệ của cấu hình bot
func validateBotConfig(config *cgbdb.BotConfig) error {
	if config.GameCode == "" {
		return runtime.NewError("Game code is required", 3)
	}

	// Validate BotJoinRules
	for i, rule := range config.BotJoinRules {
		if rule.MinBet < 0 || rule.MaxBet < rule.MinBet {
			return runtime.NewError("Invalid bet range in bot_join_rules at index "+strconv.Itoa(i), 3)
		}
		if rule.MinUsers < 1 || rule.MaxUsers < rule.MinUsers {
			return runtime.NewError("Invalid user range in bot_join_rules at index "+strconv.Itoa(i), 3)
		}
		if rule.RandomTimeMin < 0 || rule.RandomTimeMax < rule.RandomTimeMin {
			return runtime.NewError("Invalid random time in bot_join_rules at index "+strconv.Itoa(i), 3)
		}
		if rule.JoinPercent < 0 || rule.JoinPercent > 100 {
			return runtime.NewError("Invalid join_percent in bot_join_rules at index "+strconv.Itoa(i), 3)
		}
	}

	// Validate BotLeaveRules
	for i, rule := range config.BotLeaveRules {
		if rule.MinBet < 0 || rule.MaxBet < rule.MinBet {
			return runtime.NewError("Invalid bet range in bot_leave_rules at index "+strconv.Itoa(i), 3)
		}
		if rule.LeavePercent < 0 || rule.LeavePercent > 100 {
			return runtime.NewError("Invalid leave_percent in bot_leave_rules at index "+strconv.Itoa(i), 3)
		}
		if rule.LastResult != 0 && rule.LastResult != 1 && rule.LastResult != -1 {
			return runtime.NewError("Invalid last_result in bot_leave_rules at index "+strconv.Itoa(i), 3)
		}
	}

	// Validate BotCreateTableRules
	for i, rule := range config.BotCreateTableRules {
		if rule.MinBet < 0 || rule.MaxBet < rule.MinBet {
			return runtime.NewError("Invalid bet range in bot_create_table_rules at index "+strconv.Itoa(i), 3)
		}
		if rule.MinActiveTables < 0 || rule.MaxActiveTables < rule.MinActiveTables {
			return runtime.NewError("Invalid active_tables in bot_create_table_rules at index "+strconv.Itoa(i), 3)
		}
		if rule.WaitTimeMin < 0 || rule.WaitTimeMax < rule.WaitTimeMin {
			return runtime.NewError("Invalid wait_time in bot_create_table_rules at index "+strconv.Itoa(i), 3)
		}
		if rule.RetryWaitMin < 0 || rule.RetryWaitMax < rule.RetryWaitMin {
			return runtime.NewError("Invalid retry_wait in bot_create_table_rules at index "+strconv.Itoa(i), 3)
		}
	}

	// Validate BotGroupRules
	for i, rule := range config.BotGroupRules {
		if rule.VIPMin < 0 || rule.VIPMax < rule.VIPMin {
			return runtime.NewError("Invalid VIP range in bot_group_rules at index "+strconv.Itoa(i), 3)
		}
		if rule.MCBMin < 0 || rule.MCBMax < rule.MCBMin {
			return runtime.NewError("Invalid MCB range in bot_group_rules at index "+strconv.Itoa(i), 3)
		}
	}

	return nil
}
