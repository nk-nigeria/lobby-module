package api

import (
	"context"
	"github.com/ciaolink-game-platform/cgp-lobby-module/constant"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

func InitLeaderBoard(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, unmarshaler *protojson.UnmarshalOptions) {
	objectIds := []*runtime.StorageRead{
		{
			Collection: kLobbyCollection,
			Key:        kGameKey,
		},
	}
	objects, err := nk.StorageRead(ctx, objectIds)
	if err != nil {
		logger.Error("Error when read list game, error %s", err.Error())
		return
	}
	if len(objects) == 0 {
		logger.Error("Empty list game in storage ")
		return
	}
	gameListResponse := &pb.GameListResponse{}
	err = unmarshaler.Unmarshal([]byte(objects[0].GetValue()), gameListResponse)
	if err != nil {
		logger.Debug("Can not unmarshaler list game for collection")
		return
	}

	for _, game := range gameListResponse.Games {
		authoritative := false
		sort := "desc"
		operator := "incr"
		reset := constant.RESET_SCHEDULER_LEADER_BOARD
		metadata := map[string]interface{}{}
		if err := nk.LeaderboardCreate(ctx, game.Code, authoritative, sort, operator, reset, metadata); err != nil {
			logger.Debug("Can not create ")
		}
	}
}
