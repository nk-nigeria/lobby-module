package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	pb "github.com/nakamaFramework/cgp-common/proto"
	"google.golang.org/protobuf/encoding/protojson"
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

func RpcGetQuickChat(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}

		account, err := nk.AccountGetId(ctx, userID)
		if err != nil {
			return "", err
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(account.User.GetMetadata()), &metadata); err != nil {
			return "", errors.New("Corrupted user metadata.")
		}

		qcMeta := metadata["qc"]
		resData := &pb.QuickChatResponse{}
		if qcMeta != nil {
			metaText := qcMeta.([]interface{})
			for _, text := range metaText {
				resData.Texts = append(resData.Texts, text.(string))
			}
		}

		marshaler.EmitUnpopulated = true
		res, err := marshaler.Marshal(resData)
		if err != nil {
			return "", fmt.Errorf("Marharl texts error: %s", err.Error())
		}

		return string(res), nil
	}
}
