package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	objectstorage "github.com/ciaolink-game-platform/cgp-lobby-module/object-storage"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
)

func RpcGetProfile(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}

		profile, err := GetProfileUser(ctx, nk, userID)
		if err != nil {
			logger.Error("GetProfileUser error: %s", err.Error())
			return "", err
		}
		dataString, err := marshaler.Marshal(profile)
		if err != nil {
			return "", fmt.Errorf("Marharl profile error: %s", err.Error())
		}
		return string(dataString), nil
	}
}

func RpcUpdateProfile(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions, objStorage objectstorage.ObjStorage) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}
		profile := &pb.Profile{}
		if err := unmarshaler.Unmarshal([]byte(payload), profile); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		metadata := make(map[string]interface{}, 0)
		if profile.Status != "" {
			metadata["status"] = profile.Status
		}
		if profile.RefCode != "" {
			metadata["ref_code"] = profile.RefCode
		}
		if profile.AppConfig != "" {
			metadata["app_config"] = profile.AppConfig
		}
		avatarFileName := profile.GetAvatarUrl()
		avatarPresignGet := ""
		if avatarFileName != "" {
			expiry := 6 * 24 * time.Hour
			avatarPresignGet, _ = objStorage.PresignGetObject(entity.BucketAvatar, avatarFileName, expiry, nil)
		}
		err := nk.AccountUpdateId(ctx, userID, "", metadata, "", "", "", profile.LangTag, avatarPresignGet)
		if err != nil {
			logger.Error("Update userid %s error: %s", userID, err.Error())
			return "", err
		}

		newProfile, err := GetProfileUser(ctx, nk, userID)
		dataString, err := marshaler.Marshal(newProfile)
		if err != nil {
			return "", fmt.Errorf("Marharl profile error: %s", err.Error())
		}
		return string(dataString), nil
	}
}

func RpcUpdatePassword(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		// userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		// if !ok {
		// 	return "", errors.New("Missing user ID.")
		// }
		// // customUser := &entity.CustomUser{}
		// if err := json.Unmarshal([]byte(payload), customUser); err != nil {
		// 	return "", presenter.ErrUnmarshal
		// }
		// todo update user
		return "", nil
	}
}

func RpcUploadAvatar(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions, objStorage objectstorage.ObjStorage) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}
		objName := fmt.Sprintf(entity.AvatarFileName, userID)
		presignUrl, err := objStorage.PresigPutObject(entity.BucketAvatar, objName, 1*time.Hour, nil)
		if err != nil {
			logger.Error("Presign put avatar url failed, error: %s", err.Error())
			return "", err
		}
		profile := pb.Profile{
			UserId:    userID,
			AvatarUrl: presignUrl,
		}
		dataString, err := marshaler.Marshal(&profile)
		if err != nil {
			return "", fmt.Errorf("Marharl profile error: %s", err.Error())
		}
		return string(dataString), nil
	}
}

func GetProfileUser(ctx context.Context, nk runtime.NakamaModule, userID string) (*pb.Profile, error) {
	accounts, err := nk.UsersGetId(ctx, []string{userID}, nil)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, errors.New("List account empty.")
	}
	account := accounts[0]
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(account.GetMetadata()), &metadata); err != nil {
		return nil, errors.New("Corrupted user metadata.")
	}

	// todo read account chip, bank chip
	profile := pb.Profile{
		UserId:    account.GetId(),
		UserName:  account.GetUsername(),
		AvatarUrl: account.GetAvatarUrl(),
		LangTag:   account.GetLangTag(),
		Status:    entity.InterfaceToString(metadata["status"]),
		RefCode:   entity.InterfaceToString(metadata["ref_code"]),
		AppConfig: entity.InterfaceToString(metadata["app_config"]),
	}
	return &profile, nil
}
