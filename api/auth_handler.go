package api

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	"github.com/nakamaFramework/cgb-lobby-module/cgbdb"
	pb "github.com/nakamaFramework/cgp-common/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func RpcUserChangePass(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}
		request := &pb.ChangePasswordRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			logger.Error("unmarshal create match error %v", err)
			return "", presenter.ErrUnmarshal
		}
		err := cgbdb.ChangePasswordUser(ctx, logger, db,
			userId, request.GetOldPassword(), request.GetPassword())
		if err != nil {
			logger.Error("Change password user %s, error: %s", userId, err.Error())
			return "", err
		}
		return "", nil
	}
}

func RpcLinkUsername(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, _ string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		logger.Info("request link username")
		// parse request
		userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			logger.Error("request link username userid error", ok)
			return "", presenter.ErrNoUserIdFound
		}
		request := &pb.RegisterRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			logger.Error("unmarshal register request error %v", err)
			return "", presenter.ErrUnmarshal
		}

		logger.Info("user %s request register %v", userId, request)
		err := cgbdb.LinkUsername(ctx, logger, db, userId, request.UserName, request.Password)
		if err != nil {
			logger.Error("link username error", err)
		}

		return "", err
	}
}
