package api

import (
	"context"
	"database/sql"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	pb "github.com/nakama-nigeria/cgp-common/proto"
	"github.com/nakama-nigeria/lobby-module/api/presenter"
	objectstorage "github.com/nakama-nigeria/lobby-module/object-storage"
	"google.golang.org/protobuf/proto"
)

func RpcGetPreSignPut(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions, wrapper objectstorage.ObjStorage) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		preSignPutRequest := &pb.PreSignPutRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), preSignPutRequest); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		u, err := wrapper.PresigPutObject(preSignPutRequest.BucketName, preSignPutRequest.GetFileName(), 60*time.Second, nil)
		if err != nil {
			logger.Error("Error create PresigPutObject", err.Error())
			return "", presenter.ErrInternalError
		}
		preSignPutRepose := &pb.PreSignPutResponse{
			Url: u,
		}
		preSignPutReposeStr, _ := marshaler.Marshal(preSignPutRepose)
		return string(preSignPutReposeStr), nil
	}
}
