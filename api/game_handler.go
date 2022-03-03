package api

import (
	"context"
	"database/sql"

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
			return "", err
		}
		logger.Info(objects[0].GetValue())
		return objects[0].GetValue(), nil
	}
}
