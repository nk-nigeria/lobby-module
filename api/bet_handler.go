package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	"github.com/nakamaFramework/cgb-lobby-module/cgbdb"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	"github.com/nakamaFramework/cgb-lobby-module/entity"
	"github.com/nakamaFramework/cgp-common/define"
	pb "github.com/nakamaFramework/cgp-common/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// 1. bank (ẩn vip 0 1) > vip 0 1 vẫn hiện bank nhưng khi bấm vào có
//   thông báo vip 2 trở lên ms đc dùng hoặc có 1 nhãn vip 2 và k click vào đc
// 2. vip 0 vào được sảnh chọn mcb (hiện đầy đủ), chỉ chọn đc mcb thấp nhất, những mcb khác tối màu (k ẩn)
// 3.vip 0,1 k tạo được bàn > thêm 1 thông báo khi click vào: vip 2 trở lên ms tạo đc bàn
// 4.các mcb trong tab chọn mcb và tab chọn bàn hiển thị đầy đủ vs các user
//    + user k đủ tiền bấm vào sẽ hiện thông báo k đủ tiền và có 1 btn trỏ tới shop
//    + user vip 0 1 > vẫn hiển thị mcb cao nhất > click vào hiện thông báo dành cho vip 2 trở lên
//    + user vip 2 trở lên > vẫn hiện mcb thấp nhất > click vào hiện thông báo dành cho user vip 0 1
// 5.vip farm: hiện cho all vip> có dán nhãn vip 2 hoặc click vào có thông báo trên vip 2 ms đc dùng

const (
	kBetsCollection  = "bets"
	kChinesePokerKey = "chinese-poker"
)

func RpcBetList(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			logger.Error("context did not contain user ID.")
			return "", presenter.ErrInternalError
		}
		request := &pb.BetListRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			logger.Error("RpcBetList Unmarshal payload error: %s", err.Error())
			return "", presenter.ErrUnmarshal
		}
		quickJoin := false
		msg, err := loadBetsForUser(ctx, logger, db, nk, request.Code, quickJoin, userID)
		if err != nil {
			return "", err
		}
		betsJson, _ := marshaler.Marshal(msg)
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
		mapBetsByGameCode.Delete(strconv.Itoa(bet.GameId))
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
		mapBetsByGameCode.Delete(strconv.Itoa(bet.GameId))
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
		req := &pb.Bet{}
		if err := conf.Unmarshaler.Unmarshal([]byte(payload), req); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		if req.Id <= 0 {
			logger.Error("Missing bet id")
			return "", presenter.ErrNoInputAllowed
		}
		bet, err := cgbdb.ReadBet(ctx, db, req.Id)
		if err != nil {
			logger.WithField("err", err).Error("read bet failed")
			return "", presenter.ErrNotFound
		}
		betDeleted, err := cgbdb.DeleteBet(ctx, db, bet.Id)
		if betDeleted != nil {
			mapBetsByGameCode.Delete(strconv.Itoa(bet.GameId))
		}
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

func LoadBets(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, gameCode string) ([]entity.Bet, error) {
	var game entity.Game
	{
		v, exist := mapGameByCode.Load(gameCode)
		if !exist {
			cacheListGame(ctx, db, logger)
			v, exist = mapGameByCode.Load(gameCode)
			if !exist {
				return nil, fmt.Errorf("not found game id from game code %s", gameCode)
			}
		}
		game = v.(entity.Game)
	}
	values, exist := mapBetsByGameCode.Load(game.LobbyId)
	if exist {
		// return nil, fmt.Errorf("not found game id from game code %s", gameCode)
		ml := values.([]entity.Bet)
		response := make([]entity.Bet, len(ml))
		copy(response, ml)
		return response, nil
	}
	query := ""
	args := make([]interface{}, 0)
	query += "game_id=?"

	args = append(args, game.LobbyId)
	bets, _, err := cgbdb.QueryBet(ctx, db, 1000, 0, query, args...)
	if err != nil {
		logger.Error("Error when unmarshal list bets, error %s", err.Error())
		return nil, presenter.ErrInternalError
	}
	// bets := make([]entity.Bet, 0)
	// sort asc by mark unit
	sort.Slice(bets, func(i, j int) bool {
		x := bets[i]
		y := bets[j]
		return x.MarkUnit < y.MarkUnit
	})
	for idx, v := range bets {
		x := v
		x.Enable = true
		x.MaxVip = 100
		if idx == 0 {
			x.MaxVip = 1
		} else {
			x.MinVip = 1
		}
		if idx == len(bets)-1 {
			x.MinVip = 2
		}
		bets[idx] = x
	}
	mapBetsByGameCode.Store(game.LobbyId, bets)
	response := make([]entity.Bet, len(bets))
	copy(response, bets)
	return response, nil
}

func slotsGameBetConfig(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (*pb.Bets, error) {
	betsValue := []int{100, 200, 500, 1000}
	bets := make([]*pb.Bet, 0)
	for _, val := range betsValue {
		bets = append(bets, &pb.Bet{
			Enable:   true,
			MarkUnit: float32(val),
		})
	}
	userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	wallet, err := entity.ReadWalletUser(ctx, nk, logger, userID)
	if err != nil {
		logger.WithField("user", userID).WithField("err", err).Error("read wallet failed")
		return nil, err
	}
	msg := &pb.Bets{
		Bets: bets,
	}
	for _, v := range bets {
		if v.MarkUnit < float32(wallet.Chips/50) {
			msg.BestChoice = &pb.Bet{
				Enable:   true,
				MarkUnit: v.MarkUnit,
			}
		}
	}
	return msg, nil
}

func loadBetsForUser(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, gameCode string, quickJoin bool, userID string) (*pb.Bets, error) {
	if define.IsSlotGame(define.GameName(gameCode)) {
		return slotsGameBetConfig(ctx, logger, db, nk)
	}
	bets, err := LoadBets(ctx, logger, db, nk, gameCode)
	if err != nil {
		logger.WithField("err", err).Error("load bets failed")
		return nil, err
	}
	if len(bets) == 0 {
		return &pb.Bets{Bets: []*pb.Bet{}}, nil
	}
	account, _, err := cgbdb.GetProfileUser(ctx, db, userID, nil)
	if err != nil {
		logger.Error("Error when read user account error %s", err.Error())
		return nil, err
	}
	vipLv := int(account.VipLevel)
	for idx, bet := range bets {
		if vipLv > bet.MaxVip {
			bet.Enable = false
			bet.BetDisableType = pb.BetDisableType_BET_DISABLE_TYPE_ABOVE_MAX_VIP
		}
		if vipLv < bet.MinVip {
			bet.Enable = false
			bet.BetDisableType = pb.BetDisableType_BET_DISABLE_TYPE_BELOW_MIN_VIP
		}
		bets[idx] = bet
	}
	userChip := account.AccountChip
	msg := &pb.Bets{}
	trackGame := trackUserInGame[gameCode]
	for idx, bet := range bets {
		if bet.Enable {
			chipRequire := bet.AGJoin
			if quickJoin {
				chipRequire = bet.AGPlaynow
			}
			if userChip < int64(chipRequire) {
				bet.Enable = false
				bet.BetDisableType = pb.BetDisableType_BET_DISABLE_TYPE_NOT_ENOUGH_CHIP
			} else {
				bet.Enable = true
			}
		}
		bet.CountPlaying = trackGame.TotalPlayer(bet.MarkUnit)
		msg.Bets = append(msg.Bets, bet.ToPb())
		bets[idx] = bet
	}
	// best choice = first max bet enable
	for i := len(bets) - 1; i >= 0; i-- {
		bet := bets[i]
		if !bet.Enable {
			continue
		}
		msg.BestChoice = bet.ToPb()
		break
	}
	if msg.BestChoice == nil {
		msg.BestChoice = bets[len(bets)-1].ToPb()
	}
	return msg, nil
}
