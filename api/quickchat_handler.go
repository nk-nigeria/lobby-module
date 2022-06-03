package api

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
	"time"
)

func RpcUpdateQuickChat(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}

		updateReq := &pb.QuickChatUpdateRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), updateReq); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}

		ctx2, _ := context.WithTimeout(ctx, 1*time.Second)
		query := `UPDATE
					users AS u
				SET
					metadata
						= u.metadata
						|| jsonb_build_object('qc', $1::text[])
				WHERE	
					id = $2;`

		_, err := db.ExecContext(ctx2, query, updateReq.GetTexts(), userID)
		if err != nil && err != context.DeadlineExceeded {
			logger.WithField("err", err).Error("db.ExecContext last online update error.")
			return "", err
		}

		return "", nil
	}
}
