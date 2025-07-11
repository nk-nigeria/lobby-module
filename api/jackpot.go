package api

import (
	"context"
	"database/sql"

	pb "github.com/nk-nigeria/cgp-common/proto"
	"go.uber.org/zap"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nk-nigeria/lobby-module/api/presenter"
	"github.com/nk-nigeria/lobby-module/cgbdb"
	"github.com/nk-nigeria/lobby-module/conf"
)

func RpcJackpot() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		req := &pb.Jackpot{}
		if err := conf.Unmarshaler.Unmarshal([]byte(payload), req); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		if len(req.GameCode) == 0 {
			return "", presenter.ErrInvalidInput
		}
		jp, err := cgbdb.GetJackpot(ctx, logger, db, req.GameCode)
		if err != nil {
			zap.L().With(zap.Error(err)).With(zap.String("code", req.GameCode)).Error("Error when query jackpots")
			return "", err
		}
		dataJson, _ := conf.MarshalerDefault.Marshal(jp)
		return string(dataJson), nil
	}
}
