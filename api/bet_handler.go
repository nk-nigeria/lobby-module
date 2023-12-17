package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ciaolink-game-platform/cgb-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgb-lobby-module/cgbdb"
	"github.com/ciaolink-game-platform/cgb-lobby-module/conf"
	"github.com/ciaolink-game-platform/cgb-lobby-module/entity"
	pb "github.com/ciaolink-game-platform/cgp-common/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	kBetsCollection  = "bets"
	kChinesePokerKey = "chinese-poker"
)

func RpcBetList(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		request := &pb.BetListRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			logger.Error("RpcBetList Unmarshal payload error: %s", err.Error())
			return "", presenter.ErrUnmarshal
		}

		bets, err := LoadBets(request.Code, ctx, logger, db, nk)
		if err != nil {
			logger.WithField("err", err).Error("load bets failed")
			return "", err
		}
		if bets == nil || len(bets.Bets) == 0 {
			return "", nil
		}

		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			logger.Error("context did not contain user ID.")
			return "", presenter.ErrInternalError
		}
		wallets, err := entity.ReadWalletUsers(ctx, nk, logger, userID)
		if err != nil {
			logger.Error("Error when read user wallet error %s", err.Error())
			return "", presenter.ErrInternalError
		}
		if len(wallets) == 0 {
			logger.Error("Error when read user wallet error %s", err.Error())
			return "", presenter.ErrInternalError
		}
		userChip := wallets[0].Chips

		msg := &pb.Bets{}
		for _, bet := range bets.Bets {
			bet.Enable = true
			if userChip < int64(bet.AGJoin) {
				bet.Enable = false
			} else {
				bet.Enable = true
			}
			msg.Bets = append(msg.Bets, bet.ToPb())
		}
		betsJson, _ := marshaler.Marshal(msg)
		logger.Info("bets results %s", betsJson)
		return string(betsJson), nil
	}
}

// amdmin
func RpcAdminAddBet(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unauth")
		}
		bet := &entity.Bet{}
		if err := json.Unmarshal([]byte(payload), bet); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		err := cgbdb.AddBet(ctx, db, bet)
		return "", err
	}
}

func RpcAdminUpdateBet(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unauth")
		}
		bet := &entity.Bet{}
		if err := json.Unmarshal([]byte(payload), bet); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		if bet.Id <= 0 {
			logger.Error("Missing bet id")
			return "", presenter.ErrNoInputAllowed
		}
		err := cgbdb.UpdateBet(ctx, db, bet)
		if err != nil {
			logger.Error("Error when update bet, err: ", err.Error())
			return "", presenter.ErrInternalError
		}
		newBet, err := cgbdb.ReadBet(ctx, db, bet.Id)
		if err != nil {
			logger.Error("Error when read bet, err: ", err.Error())
			return "", presenter.ErrInternalError
		}
		dataStr, _ := json.Marshal(newBet)
		return string(dataStr), nil
	}
}

func RpcAdminDeleteBet(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unauth")
		}
		bet := &pb.Bet{}
		if err := conf.Unmarshaler.Unmarshal([]byte(payload), bet); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		if bet.Id <= 0 {
			logger.Error("Missing bet id")
			return "", presenter.ErrNoInputAllowed
		}
		err := cgbdb.DeleteBet(ctx, db, bet.Id)
		return "", err
	}
}

func RpcAdminListBet(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unauth")
		}
		req := &pb.BetRequest{}
		if err := conf.Unmarshaler.Unmarshal([]byte(payload), req); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		query := ""
		args := make([]interface{}, 0)
		if req.GameId > 0 {
			query += "game_id=?"
			args = append(args, req.GameId)
		}
		offset := max(0, req.Offset)
		limit := max(0, req.Limit)
		ml, total, err := cgbdb.QueryBet(ctx, db, limit, offset, query, args...)
		if err != nil {
			return "", err
		}
		res := &pb.Bets{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		}
		for _, bet := range ml {
			res.Bets = append(res.Bets, bet.ToPb())
		}
		dataStr, _ := conf.MarshalerDefault.Marshal(res)
		return string(dataStr), nil
	}
}

func LoadBets(code string, ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (*entity.ListBets, error) {
	query := ""
	args := make([]interface{}, 0)
	query += "game_id=?"
	game, exist := mapGameByCode[code]
	if !exist {
		return nil, fmt.Errorf("not found game id from game code %s", code)
	}
	args = append(args, game.LobbyId)
	bets, _, err := cgbdb.QueryBet(ctx, db, 1000, 0, query, args...)
	if err != nil {
		logger.Error("Error when unmarshal list bets, error %s", err.Error())
		return nil, presenter.ErrInternalError
	}
	return &entity.ListBets{Bets: bets}, nil

}
