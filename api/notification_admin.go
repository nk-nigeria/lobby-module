package api

import (
	"context"
	"database/sql"
	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgp-lobby-module/cgbdb"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

func RpcAddNotification(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		request := &pb.Notification{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			logger.Error("unmarshal notification error %v", err)
			return "", presenter.ErrUnmarshal
		}
		request.SenderId = ""
		err := cgbdb.AddNotification(ctx, logger, db, nk, request)
		if err != nil {
			logger.Error("Add notification user %s, error: %s", request.SenderId, err.Error())
			return "", err
		}
		return "success", nil
	}
}
