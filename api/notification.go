package api

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
	pb "github.com/nk-nigeria/cgp-common/proto"
	"github.com/nk-nigeria/lobby-module/api/presenter"
	"github.com/nk-nigeria/lobby-module/cgbdb"
	"github.com/nk-nigeria/lobby-module/conf"
	"google.golang.org/protobuf/proto"
)

func RpcReadAllNotification(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, _ string) (string, error) {
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

func RpcDeleteAllNotification(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, _ string) (string, error) {
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

func RpcReadNotification(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, _ string) (string, error) {
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

func RpcDeleteNotification(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, _ string) (string, error) {
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

func RpcListNotification(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
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
