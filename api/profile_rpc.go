package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	objectstorage "github.com/ciaolink-game-platform/cgp-lobby-module/object-storage"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
)

const DefaultLevel = 0

func RpcGetProfile(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions, objStorage objectstorage.ObjStorage) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}

		profile, _, err := GetProfileUser(ctx, nk, userID, objStorage)
		if err != nil {
			logger.Error("GetProfileUser error: %s", err.Error())
			return "", err
		}

		marshaler.EmitUnpopulated = true
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
		// metadata := make(map[string]interface{}, 0)
		_, metadata, err := GetProfileUser(ctx, nk, userID, objStorage)
		if err != nil {
			logger.Error("get profile user %s, error: %s", userID, err.Error())
			return "", err
		}
		// metadata = currentProfile.GetUserName()
		if profile.Status != "" {
			str := profile.Status
			// max len status is 255 character
			if len(str) > 256 {
				str = str[:255]
			}
			metadata["status"] = profile.Status
		}
		if profile.RefCode != "" {
			metadata["ref_code"] = profile.RefCode
		}
		if profile.AppConfig != "" {
			metadata["app_config"] = profile.AppConfig
		}
		if profile.AvatarId != "" {
			metadata["avatar_id"] = profile.AvatarId
		}
		if profile.LastOnlineTimeUnix > 0 {
			metadata["last_online_time_unix"] = profile.LastOnlineTimeUnix
		}
		// avatarFileName := profile.GetAvatarUrl()
		// avatarPresignGet := ""
		// if avatarFileName != "" {
		// 	expiry := 6 * 24 * time.Hour
		// 	var err error
		// 	avatarPresignGet, err = objStorage.PresignGetObject(entity.BucketAvatar, avatarFileName, expiry, nil)
		// 	if err != nil {
		// 		logger.Error("Presgin Avatar url failed:", err.Error())
		// 	}
		// }
		err = nk.AccountUpdateId(ctx, userID, "", metadata, profile.GetUserName(), "", "", profile.LangTag, profile.AvatarUrl)
		if err != nil {
			logger.Error("Update userid %s error: %s", userID, err.Error())
			return "", err
		}

		newProfile, _, err := GetProfileUser(ctx, nk, userID, objStorage)
		marshaler.EmitUnpopulated = true
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
		profile := &pb.Profile{}
		if err := unmarshaler.Unmarshal([]byte(payload), profile); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		if profile.AvatarUrl == "" {
			return "", presenter.ErrInternalError
		}
		// objName := fmt.Sprintf(entity.AvatarFileName, userID)
		objName := profile.AvatarUrl
		presignUrl, err := objStorage.PresigPutObject(entity.BucketAvatar, objName, 1*time.Hour, nil)
		if err != nil {
			logger.Error("Presign put avatar url failed, error: %s", err.Error())
			return "", err
		}
		profile = &pb.Profile{
			UserId:    userID,
			AvatarUrl: presignUrl,
		}
		dataString, err := marshaler.Marshal(profile)
		if err != nil {
			return "", fmt.Errorf("Marharl profile error: %s", err.Error())
		}
		return string(dataString), nil
	}
}

func GetProfileUser(ctx context.Context, nk runtime.NakamaModule, userID string, objStorage objectstorage.ObjStorage) (*pb.Profile, map[string]interface{}, error) {
	account, err := nk.AccountGetId(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(account.User.GetMetadata()), &metadata); err != nil {
		return nil, nil, errors.New("Corrupted user metadata.")
	}

	user := account.User
	// todo read account chip, bank chip
	profile := pb.Profile{
		UserId:             user.GetId(),
		UserName:           user.GetUsername(),
		LangTag:            user.GetLangTag(),
		DisplayName:        user.GetDisplayName(),
		Status:             entity.InterfaceToString(metadata["status"]),
		RefCode:            entity.InterfaceToString(metadata["ref_code"]),
		AppConfig:          entity.InterfaceToString(metadata["app_config"]),
		LinkGroup:          entity.LinkGroupFB,
		LinkFanpageFb:      entity.LinkFanpageFB,
		AvatarId:           entity.InterfaceToString(metadata["avatar_id"]),
		VipLevel:           entity.ToInt64(metadata["vip_level"], DefaultLevel),
		LastOnlineTimeUnix: entity.ToInt64(metadata["last_online_time_unix"], 0),
	}

	if profile.DisplayName == "" {
		profile.DisplayName = profile.UserName
	}

	if strings.HasPrefix(profile.UserName, entity.AutoPrefix) {
		profile.Registrable = true
	} else {
		profile.Registrable = false
	}

	if user.GetAvatarUrl() != "" && objStorage != nil {
		// objName := fmt.Sprintf(entity.AvatarFileName, userID)
		objName := user.GetAvatarUrl()
		avatatUrl, _ := objStorage.PresignGetObject(entity.BucketAvatar, objName, 24*time.Hour, nil)
		profile.AvatarUrl = avatatUrl
	}

	if account.GetWallet() != "" {
		wallet, err := entity.ParseWallet(account.GetWallet())
		if err == nil {
			profile.AccountChip = wallet.Chips
			profile.BankChip = wallet.ChipsInBank
		}
	}

	return &profile, metadata, nil
}
