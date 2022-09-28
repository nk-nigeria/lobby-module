package api

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/ciaolink-game-platform/cgb-lobby-module/entity"

	"github.com/ciaolink-game-platform/cgb-lobby-module/api/presenter"
	pb "github.com/ciaolink-game-platform/cgb-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	kLobbyCollection = "lobby"
	kGameKey         = "games"
)

func InitListGame(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) {
	objectIds := []*runtime.StorageRead{
		{
			Collection: kLobbyCollection,
			Key:        kGameKey,
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
	var games entity.Games
	games.List = []entity.Game{
		{
			Code: "noel-slot",
			Layout: entity.Layout{
				Col:     1,
				Row:     1,
				ColSpan: 2,
				RowSpan: 2,
			},
			LobbyId: "2",
		},

		{
			Code: "color-game",
			Layout: entity.Layout{
				Col:     1,
				Row:     2,
				ColSpan: 2,
				RowSpan: 2,
			},
			LobbyId: "2",
		},

		{
			Code: "roulette",
			Layout: entity.Layout{
				Col:     1,
				Row:     3,
				ColSpan: 2,
				RowSpan: 2,
			},
			LobbyId: "4",
		},

		{
			Code: "fruit-slot",
			Layout: entity.Layout{
				Col:     1,
				Row:     4,
				ColSpan: 2,
				RowSpan: 2,
			},
			LobbyId: "5",
		},

		{
			Code: "sabong-cards",
			Layout: entity.Layout{
				Col:     2,
				Row:     1,
				ColSpan: 2,
				RowSpan: 2,
			},
			LobbyId: "6",
		},

		{
			Code: "chinese-poker",
			Layout: entity.Layout{
				Col:     2,
				Row:     2,
				ColSpan: 2,
				RowSpan: 2,
			},
			LobbyId: "7",
		},

		{
			Code: "baccarat",
			Layout: entity.Layout{
				Col:     2,
				Row:     3,
				ColSpan: 2,
				RowSpan: 2,
			},
			LobbyId: "8",
		},

		{
			Code: "lucky-number",
			Layout: entity.Layout{
				Col:     2,
				Row:     4,
				ColSpan: 2,
				RowSpan: 2,
			},
			LobbyId: "9",
		},
	}

	gameJson, err := json.Marshal(&pb.GameListResponse{
		Games: games.ToPB(),
	})
	if err != nil {
		logger.Debug("Can not marshaler list game for collection")
		return
	}
	w := &runtime.StorageWrite{
		Collection:      kLobbyCollection,
		Key:             kGameKey,
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

func RpcGameList(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		objectIds := []*runtime.StorageRead{
			{
				Collection: kLobbyCollection,
				Key:        kGameKey,
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

		return objects[0].GetValue(), nil
	}
}
