package api

import (
	"context"
	"database/sql"

	pb "github.com/nk-nigeria/cgp-common/proto"
	"github.com/nk-nigeria/lobby-module/api/presenter"
	"github.com/nk-nigeria/lobby-module/cgbdb"
	"github.com/nk-nigeria/lobby-module/conf"
	objectstorage "github.com/nk-nigeria/lobby-module/object-storage"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/proto"
)

func RpcAddInAppMessage(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		inAppMessage := &pb.InAppMessage{}
		if err := unmarshaler.Unmarshal([]byte(payload), inAppMessage); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		data, err := cgbdb.AddInAppMessage(ctx, logger, db, marshaler, inAppMessage)
		if err != nil {
			return "", err
		}
		dataStr, _ := conf.MarshalerDefault.Marshal(data)
		return string(dataStr), nil
	}
}

func RpcUpdateInAppMessage(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		inAppMessage := &pb.InAppMessage{}
		if err := unmarshaler.Unmarshal([]byte(payload), inAppMessage); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		var err error
		inAppMessage, err = cgbdb.UpdateInAppMessage(ctx, logger, db, marshaler, unmarshaler, inAppMessage.Id, inAppMessage)
		if err != nil {
			return "", err
		}
		inAppMessageStr, _ := conf.MarshalerDefault.Marshal(inAppMessage)
		return string(inAppMessageStr), nil
	}
}

func RpcListInAppMessage(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions, wrapper objectstorage.ObjStorage) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		inAppMessageRequest := &pb.InAppMessageRequest{}
		if payload != "" {
			if err := unmarshaler.Unmarshal([]byte(payload), inAppMessageRequest); err != nil {
				logger.Error("Error when unmarshal payload", err.Error())
				return "", presenter.ErrUnmarshal
			}
		}
		list, err := cgbdb.GetListInAppMessage(ctx, logger, db, unmarshaler, nk, inAppMessageRequest.Limit, inAppMessageRequest.Cusor, inAppMessageRequest.Type)
		if err != nil {
			return "", err
		}
		listInAppMessageStr, _ := conf.MarshalerDefault.Marshal(list)
		return string(listInAppMessageStr), nil
	}
}

func RpcDeleteInAppMessage(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		inAppMessage := &pb.InAppMessage{}
		if err := unmarshaler.Unmarshal([]byte(payload), inAppMessage); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		var err error
		err = cgbdb.DeleteInAppMessage(ctx, logger, db, inAppMessage.Id)
		if err != nil {
			return "", err
		}
		return string("deleted"), nil
	}
}
