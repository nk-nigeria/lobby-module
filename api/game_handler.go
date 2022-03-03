package api

import (
	"context"
	"database/sql"
	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

func RpcGameList(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		response, err := marshaler.Marshal(&pb.GameListResponse{
			Games: []*pb.Game{
				{
					Code:   "GAME1",
					Active: true,
				},
				{
					Code:   "GAME2",
					Active: true,
				},
				{
					Code:   "GAME3",
					Active: true,
				},
			},
		})
		if err != nil {
			logger.Error("error marshaling response payload: %v", err.Error())
			return "", presenter.ErrMarshal
		}

		return string(response), nil
	}
}
