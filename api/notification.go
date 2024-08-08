package api

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	"github.com/nakamaFramework/cgb-lobby-module/cgbdb"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	pb "github.com/nakamaFramework/cgp-common/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func RpcReadAllNotification(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, _ string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		// parse request
		userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		err := cgbdb.ReadAllNotification(ctx, logger, db, userId)
		if err != nil {
			logger.Error("RpcReadAllNotification error", err)
		}

		response := &pb.DefaultResponse{
			Message: "success",
			Code:    "200",
			Status:  "success",
		}
		out, _ := conf.MarshalerDefault.Marshal(response)
		return string(out), err
	}
}

func RpcDeleteAllNotification(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, _ string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		// parse request
		userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		err := cgbdb.DeleteAllNotification(ctx, logger, db, userId)
		if err != nil {
			logger.Error("RpcDeleteAllNotification error", err)
		}
		response := &pb.DefaultResponse{
			Message: "success",
			Code:    "200",
			Status:  "success",
		}
		out, _ := conf.MarshalerDefault.Marshal(response)
		return string(out), err
	}
}

func RpcReadNotification(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, _ string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		// parse request
		userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}
		request := &pb.Notification{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			logger.Error("unmarshal notification error %v", err)
			return "", presenter.ErrUnmarshal
		}

		err := cgbdb.ReadNotification(ctx, logger, db, request.Id, userId)
		if err != nil {
			logger.Error("ReadNotification error", err)
		}

		response := &pb.DefaultResponse{
			Message: "success",
			Code:    "200",
			Status:  "success",
		}
		out, _ := conf.MarshalerDefault.Marshal(response)
		return string(out), err
	}
}

func RpcDeleteNotification(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, _ string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		// parse request
		userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}
		request := &pb.Notification{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			logger.Error("unmarshal notification error %v", err)
			return "", presenter.ErrUnmarshal
		}

		err := cgbdb.DeleteNotification(ctx, logger, db, request.Id, userId)
		if err != nil {
			logger.Error("DeleteNotification error", err)
		}
		response := &pb.DefaultResponse{
			Message: "success",
			Code:    "200",
			Status:  "success",
		}
		out, _ := conf.MarshalerDefault.Marshal(response)
		return string(out), err
	}
}

func RpcListNotification(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}
		notificationRequest := &pb.NotificationRequest{}
		if payload != "" {
			if err := unmarshaler.Unmarshal([]byte(payload), notificationRequest); err != nil {
				logger.Error("Error when unmarshal payload", err.Error())
				return "", presenter.ErrUnmarshal
			}
		}
		list, err := cgbdb.GetListNotification(ctx, logger, db, notificationRequest.Limit, notificationRequest.Cusor, userId, notificationRequest.Type)
		if err != nil {
			return "", err
		}
		listNotificationStr, _ := conf.MarshalerDefault.Marshal(list)
		return string(listNotificationStr), nil
	}
}
