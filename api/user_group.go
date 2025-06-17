package api

import (
	"context"
	"database/sql"

	pb "github.com/nk-nigeria/cgp-common/proto"
	"github.com/nk-nigeria/lobby-module/api/presenter"
	"github.com/nk-nigeria/lobby-module/cgbdb"
	"github.com/nk-nigeria/lobby-module/conf"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/proto"
)

func RpcAddUserGroup(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userGroup := &pb.UserGroup{}
		if err := unmarshaler.Unmarshal([]byte(payload), userGroup); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		err := cgbdb.AddUserGroup(ctx, logger, db, userGroup, marshaler)
		if err != nil {
			return "", err
		}
		userGroupStr, _ := conf.MarshalerDefault.Marshal(userGroup)
		return string(userGroupStr), nil
	}
}

func RpcUpdateUserGroup(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userGroup := &pb.UserGroup{}
		if err := unmarshaler.Unmarshal([]byte(payload), userGroup); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		var err error
		userGroup, err = cgbdb.UpdateUserGroup(ctx, logger, db, marshaler, unmarshaler, userGroup.Id, userGroup)
		if err != nil {
			return "", err
		}
		userGroupStr, _ := conf.MarshalerDefault.Marshal(userGroup)
		return string(userGroupStr), nil
	}
}

func RpcListUserGroup(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userGroupRequest := &pb.UserGroupRequest{}
		if payload != "" {
			if err := unmarshaler.Unmarshal([]byte(payload), userGroupRequest); err != nil {
				logger.Error("Error when unmarshal payload", err.Error())
				return "", presenter.ErrUnmarshal
			}
		}
		list, err := cgbdb.GetListUserGroup(ctx, logger, db, unmarshaler, userGroupRequest.Limit, userGroupRequest.Cusor)
		if err != nil {
			return "", err
		}
		listUserGroupStr, _ := conf.MarshalerDefault.Marshal(list)
		return string(listUserGroupStr), nil
	}
}

func RpcDeleteUserGroup(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userGroup := &pb.UserGroup{}
		if err := unmarshaler.Unmarshal([]byte(payload), userGroup); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		var err error
		err = cgbdb.DeleteUserGroup(ctx, logger, db, unmarshaler, userGroup.Id)
		if err != nil {
			return "", err
		}
		return string("deleted"), nil
	}
}
