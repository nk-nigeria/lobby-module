package api

import (
	"context"
	"database/sql"

	pb "github.com/nakamaFramework/cgp-common/proto"
	"go.uber.org/zap"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	"github.com/nakamaFramework/cgb-lobby-module/cgbdb"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
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
