package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api"
	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	objectstorage "github.com/ciaolink-game-platform/cgp-lobby-module/object-storage"
)

const (
	rpcIdGameList    = "list_game"
	rpcIdFindMatch   = "find_match"
	rpcIdCreateMatch = "create_match"

	rpcIdListBet = "list_bet"

	rpcUserChangePass = "user_change_pass"
	rpcLinkUsername   = "link_username"

	rpcGetProfile     = "get_profile"
	rpcUpdateProfile  = "update_profile"
	rpcUpdatePassword = "update_password"
	rpcUpdateAvatar   = "update_avatar"
)

//noinspection GoUnusedExportedFunction
func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	initStart := time.Now()

	marshaler := &protojson.MarshalOptions{
		UseEnumNumbers: true,
	}
	unmarshaler := &protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}

	api.InitListGame(ctx, logger, nk)
	api.InitListBet(ctx, logger, nk)

	objStorage, err := InitObjectStorage(logger)
	if err != nil {
		objStorage = &objectstorage.EmptyStorage{}
	} else {
		objStorage.MakeBucket(entity.BucketAvatar)
	}

	if err := initializer.RegisterRpc(rpcIdGameList, api.RpcGameList(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcIdFindMatch, api.RpcFindMatch(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcIdCreateMatch, api.RpcCreateMatch(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcIdListBet, api.RpcBetList(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcUserChangePass, api.RpcUserChangePass(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcLinkUsername, api.RpcLinkUsername(marshaler, unmarshaler)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(rpcGetProfile, api.RpcGetProfile(marshaler, unmarshaler, objStorage)); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(
		rpcUpdateProfile,
		api.RpcUpdateProfile(marshaler, unmarshaler, objStorage),
	); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(
		rpcUpdatePassword,
		api.RpcUpdatePassword(marshaler, unmarshaler),
	); err != nil {
		return err
	}

	if err := initializer.RegisterRpc(
		rpcUpdateAvatar,
		api.RpcUploadAvatar(marshaler, unmarshaler, objStorage),
	); err != nil {
		return err
	}

	if err := api.RegisterSessionEvents(db, nk, initializer); err != nil {
		return err
	}

	logger.Info("Plugin loaded in '%d' msec.", time.Now().Sub(initStart).Milliseconds())
	return nil
}

const (
	MinioHost      = "103.226.250.195:9000"
	MinioKey       = "minio"
	MinioAccessKey = "minioadmin"
)

func InitObjectStorage(logger runtime.Logger) (objectstorage.ObjStorage, error) {
	w, err := objectstorage.NewMinioWrapper(MinioHost, MinioKey, MinioAccessKey, false)
	if err != nil {
		logger.Error("Init Object Storage Engine Minio error: %s", err.Error())
	}
	return w, err
}
