package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

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

func InitListBet(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) {
	objectIds := []*runtime.StorageRead{
		{
			Collection: kBetsCollection,
			Key:        kChinesePokerKey,
		},
	}

	objects, err := nk.StorageRead(ctx, objectIds)
	if err != nil {
		logger.Error("Error when read chinese poker at init, error %s", err.Error())
	}

	if len(objects) > 0 {
		logger.Info("List bet chinese poker game already write in collection")
		return
	}

	bets := &entity.ListBets{
		Bets: []entity.Bet{
			{
				Enable:    true,
				MarkUnit:  100,
				Xjoin:     5,
				AGJoin:    500,
				Xplaynow:  10,
				AGPlaynow: 1000,
				Xleave:    3,
				AGLeave:   300,
				Xfee:      0,
				AGFee:     0,
				NewFee:    0,
			},

			{
				Enable:    true,
				MarkUnit:  500,
				Xjoin:     5,
				AGJoin:    2500,
				Xplaynow:  5,
				AGPlaynow: 2500,
				Xleave:    3,
				AGLeave:   1500,
				Xfee:      0,
				AGFee:     0,
				NewFee:    0,
			},

			{
				Enable:    true,
				MarkUnit:  1000,
				Xjoin:     20,
				AGJoin:    20000,
				Xplaynow:  20,
				AGPlaynow: 20000,
				Xleave:    10,
				AGLeave:   10000,
				Xfee:      20,
				AGFee:     20000,
				NewFee:    1.5,
			},

			{
				Enable:    true,
				MarkUnit:  5000,
				Xjoin:     20,
				AGJoin:    100000,
				Xplaynow:  20,
				AGPlaynow: 100000,
				Xleave:    10,
				AGLeave:   50000,
				Xfee:      20,
				AGFee:     100000,
				NewFee:    1.5,
			},

			{
				Enable:    true,
				MarkUnit:  10000,
				Xjoin:     20,
				AGJoin:    200000,
				Xplaynow:  20,
				AGPlaynow: 200000,
				Xleave:    10,
				AGLeave:   100000,
				Xfee:      20,
				AGFee:     200000,
				NewFee:    1.5,
			},

			{
				Enable:    true,
				MarkUnit:  50000,
				Xjoin:     20,
				AGJoin:    1000000,
				Xplaynow:  20,
				AGPlaynow: 1000000,
				Xleave:    10,
				AGLeave:   500000,
				Xfee:      20,
				AGFee:     1000000,
				NewFee:    1.5,
			},
		},
	}

	betsJson, err := json.Marshal(bets)
	if err != nil {
		logger.Debug("Can not marshaler list game for collection")
		return
	}

	writeObjects := []*runtime.StorageWrite{
		{
			Collection:      kBetsCollection,
			Key:             kChinesePokerKey,
			Value:           string(betsJson),
			PermissionRead:  2,
			PermissionWrite: 0,
		},
	}

	if len(writeObjects) == 0 {
		logger.Debug("Can not generate list game for collection")
		return
	}

	_, err = nk.StorageWrite(ctx, writeObjects)
	if err != nil {
		logger.Error("Write list game for collection error %s", err.Error())
	}
}

func LoadBets(code string, ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) (*entity.ListBets, error) {
	objectIds := []*runtime.StorageRead{
		{
			Collection: kBetsCollection,
			Key:        code,
		},
	}
	objects, err := nk.StorageRead(ctx, objectIds)
	bets := &entity.ListBets{}
	if err != nil {
		logger.Error("Error when read list bet, error %s", err.Error())
		return nil, presenter.ErrBetNotFound
	}
	if len(objects) == 0 {
		logger.Warn("List bet in storage empty")
		return bets, nil
	}

	json.Unmarshal([]byte(objects[0].GetValue()), bets)
	if err != nil {
		logger.Error("Error when unmarshal list bets, error %s", err.Error())
		return bets, presenter.ErrUnmarshal
	}
	return bets, nil
}

func RpcBetList(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		request := &pb.BetListRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			logger.Error("RpcBetList Unmarshal payload error: %s", err.Error())
			return "", presenter.ErrUnmarshal
		}

		bets, err := LoadBets(request.Code, ctx, logger, nk)
		if err != nil {
			logger.Error("Error when unmarshal list bets, error %s", err.Error())
			return "", presenter.ErrUnmarshal
		}

		if len(bets.Bets) == 0 {
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
