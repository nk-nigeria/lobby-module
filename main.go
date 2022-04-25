package main

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
)

const (
	rpcIdGameList    = "list_game"
	rpcIdFindMatch   = "find_match"
	rpcIdCreateMatch = "create_match"

	rpcIdListBet = "list_bet"

	rpcPushToBank        = "push_to_bank"
	rpcWithDraw          = "with_draw"
	rpcBankSendGift      = "send_gift"
	rpcWalletTransaction = "wallet_transaction"
)

//noinspection GoUnusedExportedFunction
func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	initStart := time.Now()

	marshaler := &protojson.MarshalOptions{
		UseEnumNumbers: true,
	}
	unmarshaler := &protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}
	InitListGame(marshaler, ctx, logger, nk)
	InitListBet(marshaler, ctx, logger, nk)

	if err := initializer.RegisterRpc(rpcIdGameList, api.RpcGameList(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcIdFindMatch, api.RpcFindMatch(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcIdCreateMatch, api.RpcCreateMatch(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcIdListBet, api.RpcBetList(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcPushToBank, api.RpcPushToBank(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcWithDraw, api.RpcWithDraw(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcBankSendGift, api.RpcBankSendGift(marshaler, unmarshaler)); err != nil {
		return err
	}
	if err := initializer.RegisterRpc(rpcWalletTransaction, api.RpcWalletTransaction(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := api.RegisterSessionEvents(db, nk, initializer); err != nil {
		return err
	}

	logger.Info("Plugin loaded in '%d' msec.", time.Now().Sub(initStart).Milliseconds())
	return nil
}

func InitListGame(marshaler *protojson.MarshalOptions, ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) {
	objectIds := []*runtime.StorageRead{
		{
			Collection: "lobby_1",
			Key:        "list_game",
		},
	}
	objects, err := nk.StorageRead(ctx, objectIds)
	if err != nil {
		logger.Error("Error when read list game at init, error %s", err.Error())
	}

	// check list game has write in collection
	if len(objects) > 0 {
		logger.Info("List game already write in collection")
		return
	}
	writeObjects := []*runtime.StorageWrite{}
	games := []*pb.Game{}
	for i := 1; i <= 4; i++ {
		for j := 1; j <= 4; j++ {
			game := &pb.Game{
				Code:    "GAME_" + strconv.Itoa(i*10+j),
				Active:  i%2 == 0,
				LobbyId: strconv.Itoa(i + j),
				Layout: &pb.Layout{
					Col:     int32(i),
					Row:     int32(j),
					ColSpan: 2,
					RowSpan: 2,
				},
			}
			games = append(games, game)
		}
	}
	gameJson, err := marshaler.Marshal(&pb.GameListResponse{
		Games: games,
	})
	if err != nil {
		logger.Debug("Can not marshaler list game for collection")
		return
	}
	w := &runtime.StorageWrite{
		Collection:      "lobby_1",
		Key:             "list_game",
		Value:           string(gameJson),
		PermissionRead:  2,
		PermissionWrite: 0,
	}
	writeObjects = append(writeObjects, w)
	if len(writeObjects) == 0 {
		logger.Debug("Can not generate list game for collection")
		return
	}
	_, err = nk.StorageWrite(ctx, writeObjects)
	if err != nil {
		logger.Error("Write list game for collection error %s", err.Error())
	}
}

func InitListBet(marshaler *protojson.MarshalOptions, ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) {
	objectIds := []*runtime.StorageRead{
		{
			Collection: "bets",
			Key:        "chinese-poker",
		},
	}
	objects, err := nk.StorageRead(ctx, objectIds)
	if err != nil {
		logger.Error("Error when read list game at init, error %s", err.Error())
	}

	// check list game has write in collection
	if len(objects) > 0 {
		logger.Info("List game already write in collection")
		return
	}
	writeObjects := []*runtime.StorageWrite{}
	var bets []int32
	for i := int32(1); i <= 10; i++ {
		bets = append(bets, i)
	}
	gameJson, err := marshaler.Marshal(&pb.BetListResponse{
		Bets: bets,
	})
	if err != nil {
		logger.Debug("Can not marshaler list game for collection")
		return
	}
	w := &runtime.StorageWrite{
		Collection:      "bets",
		Key:             "chinese-poker",
		Value:           string(gameJson),
		PermissionRead:  2,
		PermissionWrite: 0,
	}
	writeObjects = append(writeObjects, w)
	if len(writeObjects) == 0 {
		logger.Debug("Can not generate list game for collection")
		return
	}
	_, err = nk.StorageWrite(ctx, writeObjects)
	if err != nil {
		logger.Error("Write list game for collection error %s", err.Error())
	}
}
