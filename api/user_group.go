package api

import (
	"context"
	"database/sql"
	"strings"
	"unicode"

	"github.com/ciaolink-game-platform/cgb-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgb-lobby-module/cgbdb"
	"github.com/ciaolink-game-platform/cgb-lobby-module/conf"
	"github.com/ciaolink-game-platform/cgb-lobby-module/constant"
	pb "github.com/ciaolink-game-platform/cgb-lobby-module/proto"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

func removeSpace(s string) string {
	rr := make([]rune, 0, len(s))
	for _, r := range s {
		if !unicode.IsSpace(r) {
			rr = append(rr, r)
		}
	}
	return string(rr)
}

func RpcAddUserGroup(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userGroup := &pb.UserGroup{}
		if err := unmarshaler.Unmarshal([]byte(payload), userGroup); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		typeUG := constant.UserGroupType(userGroup.Type)
		if typeUG != constant.UserGroupType_Level &&
			typeUG != constant.UserGroupType_VipLevel &&
			typeUG != constant.UserGroupType_WalletChips &&
			typeUG != constant.UserGroupType_WalletChipsInbank &&
			typeUG != constant.UserGroupType_All {
			logger.Error("Error user group not valid")
			return "", presenter.ErrUnmarshal
		}
		userGroup.Data = removeSpace(userGroup.Data)
		userGroup.Data = strings.Trim(userGroup.Data, " ")
		err := cgbdb.AddUserGroup(ctx, logger, db, userGroup, marshaler)
		if err != nil {
			return "", err
		}
		userGroupStr, _ := conf.MarshalerDefault.Marshal(userGroup)
		return string(userGroupStr), nil
	}
}

func RpcUpdateUserGroup(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
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

func RpcListUserGroup(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
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

func RpcDeleteUserGroup(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
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
