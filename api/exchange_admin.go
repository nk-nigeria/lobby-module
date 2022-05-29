package api

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgp-lobby-module/cgbdb"
	"github.com/ciaolink-game-platform/cgp-lobby-module/conf"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
)

func RpcGetAllExchange() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unath.")
		}
		exChangedealReq := &pb.ExchangeRequest{}
		unmarshaler := conf.Unmarshaler
		if err := unmarshaler.Unmarshal([]byte(payload), exChangedealReq); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		list, err := cgbdb.GetAllExchange(ctx, logger, db, userID, exChangedealReq)
		if err != nil {
			logger.Error("Error when get all list exchange", err.Error())
			return "", presenter.ErrUnmarshal
		}
		sarshaler := conf.Marshaler
		listJson, _ := sarshaler.Marshal(list)
		return string(listJson), nil
	}
}
