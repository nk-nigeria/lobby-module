package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	"github.com/nakamaFramework/cgb-lobby-module/cgbdb"
	"github.com/nakamaFramework/cgb-lobby-module/entity"
	objectstorage "github.com/nakamaFramework/cgb-lobby-module/object-storage"
	pb "github.com/nakamaFramework/cgp-common/proto"
)

const DefaultLevel = 0

func RpcGetProfile(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions, objStorage objectstorage.ObjStorage) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}

		profile, _, err := cgbdb.GetProfileUser(ctx, db, userID, objStorage)
		if err != nil {
			logger.Error("GetProfileUser error: %s", err.Error())
			return "", err
		}
		// check match valid in profile
		if len(profile.PlayingMatch.MatchId) > 0 {
			match, err := nk.MatchGet(ctx, profile.PlayingMatch.MatchId)
			if err != nil || match == nil {
				logger.WithField("err", err).WithField("match", match).Error("MatchGet error")
				profile.PlayingMatch.MatchId = ""
				cgbdb.UpdateUsersPlayingInMatch(ctx, logger, db, userID, profile.PlayingMatch)
			}
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
		currentProfile, metadata, err := cgbdb.GetProfileUser(ctx, db, userID, objStorage)
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
		addNewReferUser := false
		if currentProfile.RemainTimeInputRefCode > 0 &&
			entity.InterfaceToString(metadata["ref_code"]) == "" {
			profile.RefCode = strings.TrimSpace(profile.RefCode)
			if profile.RefCode != "" {
				// check valid ref code
				userSidStr := strconv.Itoa(int(currentProfile.UserSid))
				if profile.RefCode == currentProfile.UserId || profile.RefCode == userSidStr {
					return "", status.Error(codes.InvalidArgument, "Can not ref yourself")
				}
				// if using user sid
				if refCodeInt, _ := strconv.Atoi(profile.RefCode); refCodeInt > 0 {
					_, err = cgbdb.GetAccount(ctx, db, "", int64(refCodeInt))
				} else {
					//  using user id (uuid)
					_, err = nk.AccountGetId(ctx, profile.RefCode)
				}
				if err != nil {
					logger.Error("Error when valid ref code %s err %s", profile.RefCode, err.Error())
					return "", status.Error(codes.InvalidArgument, "Invalid ref code")
				}
				metadata["ref_code"] = profile.RefCode
				addNewReferUser = true
			}
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

		newProfile, _, err := cgbdb.GetProfileUser(ctx, db, userID, objStorage)
		if addNewReferUser {
			userRefer := &pb.ReferUser{
				UserInvitor: profile.RefCode,
				UserInvitee: newProfile.UserId,
			}
			cgbdb.AddUserRefer(ctx, logger, db, userRefer)
		}
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
