package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

func RpcGameList(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		objectIds := []*runtime.StorageRead{
			{
				Collection: "lobby_1",
				Key:        "list_game",
			},
		}
		objects, err := nk.StorageRead(ctx, objectIds)
		if err != nil {
			logger.Error("Error when read list game, error %s", err.Error())
			return "", presenter.ErrMarshal
		}
		if len(objects) == 0 {
			logger.Error("Empty list game in storage ")
			return "", nil
		}
		queryParms, ok := ctx.Value(runtime.RUNTIME_CTX_QUERY_PARAMS).(map[string][]string)
		if !ok {
			queryParms = make(map[string][]string)
		}
		arr := queryParms["enable_filter_chip"]
		if len(arr) == 0 {
			return objects[0].GetValue(), nil
		}
		v := strings.ToLower(arr[0])
		if v != "1" && v != "true" {
			return objects[0].GetValue(), nil
		}
		games := entity.Games{}
		err = json.Unmarshal([]byte(objects[0].GetValue()), &games)
		if err != nil {
			logger.Error("Error when unmarshal list game, error %s", err.Error())
			return "", presenter.ErrUnmarshal
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
		for idx, game := range games.List {
			if game.MinChip > 0 && userChip < int64(game.MinChip) {
				game.Enable = false
			} else {
				game.Enable = true
			}
			games.List[idx] = game
		}
		gamesJson, _ := json.Marshal(games)
		return string(gamesJson), nil
	}
}

func RpcBetList(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		request := &pb.BetListRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			return "", presenter.ErrUnmarshal
		}

		bets, err := entity.LoadBets(request.Code, ctx, logger, nk, unmarshaler)
		if err != nil {
			logger.Error("Error when unmarshal list bets, error %s", err.Error())
			return "", presenter.ErrUnmarshal
		}
		if len(bets.Bets) == 0 {
			return "", nil
		}
		queryParms, ok := ctx.Value(runtime.RUNTIME_CTX_QUERY_PARAMS).(map[string][]string)
		if !ok {
			queryParms = make(map[string][]string)
		}
		arr := queryParms["enable_filter_chip"]

		v := strings.ToLower(arr[0])
		if v != "1" && v != "true" {
			return "", nil
		}
		if len(arr) == 0 {
			betsJson, _ := marshaler.Marshal(bets)
			return string(betsJson), nil
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
		for idx, bet := range bets.Bets {
			bet.Enable = true
			if userChip < int64(bet.GetAgJoin()) {
				bet.Enable = false
			} else {
				bet.Enable = true
			}
			bets.Bets[idx] = bet
		}
		betsJson, _ := marshaler.Marshal(bets)
		return string(betsJson), nil
		// return "{   \"bets\": [     100,     500,     5000,     20000,     50000,     100000,     200000,     500000,     1000000   ] }", nil
	}
}
