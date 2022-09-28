package cgbdb

import (
	"context"
	"database/sql"
	"regexp"
	"strings"

	"github.com/ciaolink-game-platform/cgb-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgb-lobby-module/constant"

	"github.com/ciaolink-game-platform/cgb-lobby-module/entity"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	invalidUsernameRegex = regexp.MustCompilePOSIX("([[:cntrl:]]|[[\t\n\r\f\v]])+")
	invalidCharsRegex    = regexp.MustCompilePOSIX("([[:cntrl:]]|[[:space:]])+")
	emailRegex           = regexp.MustCompile("^.+@.+\\..+$")
)

func ChangePasswordUser(ctx context.Context, logger runtime.Logger, db *sql.DB, userId, oldpassword, newpassword string) error {
	query := "SELECT id, password, disable_time FROM users WHERE id = $1"
	var dbUserID string
	var dbPassword []byte
	var dbDisableTime pgtype.Timestamptz

	err := db.QueryRowContext(ctx, query, userId).Scan(&dbUserID, &dbPassword, &dbDisableTime)
	if err != nil {
		logger.Error("Userid %s not found", userId)
		return status.Error(codes.Internal, "Error finding user account.")
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newpassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Error hashing new password username %s, err: %s", userId, err.Error())
		return status.Error(codes.Internal, "Error hash password user account.")
	}

	if dbDisableTime.Status == pgtype.Present && dbDisableTime.Time.Unix() != 0 {
		logger.Info("Userid %s  account is disabled.", userId)
		return status.Error(codes.PermissionDenied, "User account banned.")
	}

	if len(dbPassword) == 0 {
		logger.Error("Can't change account has password = nil")
		// Do not disambiguate between bad password and password login not possible at all in client-facing error messages.
		return status.Error(codes.Unauthenticated, "Invalid credentials.")
	}

	if bcrypt.CompareHashAndPassword(dbPassword, []byte(oldpassword)) != nil {
		logger.Error("Can't change account %s, username or old password wrong", userId)
		return status.Error(codes.Unauthenticated, "Invalid credentials.")
	}

	result, err := db.ExecContext(ctx, `
UPDATE users SET password=$1, update_time = now()
WHERE id=$2`, hashedNewPassword, userId)
	if err != nil {
		logger.Error("Cannot update password user with userId %s, err: %s", userId, err.Error())
		return status.Error(codes.Internal, "Error update password user account.")
	}

	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not update new user.")
		return status.Error(codes.Internal, "Error finding user account.")
	}

	return nil
}

func LinkUsername(ctx context.Context, logger runtime.Logger, db *sql.DB, userID, username, password string) error {
	logger.Info("begin link username")
	if username == "" || password == "" {
		return presenter.ErrUserNameAndPasswordRequired
	} else if invalidUsernameRegex.MatchString(username) {
		return presenter.ErrUserNameInvalid
	} else if len(password) < 8 {
		return presenter.ErrUserPasswordLenthTooShort
	} else {
		lUserName := len(username)
		if lUserName < 8 {
			return presenter.ErrUserNameLenthTooShort
		}
		if lUserName > 255 {
			return presenter.ErrUserNameLenthTooLong
		}
	}

	cleanUsername := strings.ToLower(username)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	res, err := db.ExecContext(ctx, `
UPDATE users
SET username = $2, password = $3, update_time = now()
WHERE (id = $1)
AND (NOT EXISTS
    (SELECT id
     FROM users
     WHERE username = $2 AND NOT id = $1))`,
		userID,
		cleanUsername,
		hashedPassword)

	if err != nil {
		logger.Error("Could not link username.", zap.Error(err), zap.Any("input", username))
		return status.Error(codes.Internal, "Error while trying to link username.")
	} else if count, _ := res.RowsAffected(); count == 0 {
		return presenter.ErrUserNameExist
	}
	return nil
}

func FetchUserIDWithCondition(ctx context.Context, db *sql.DB, condition string, params []interface{}) ([]string, error) {
	ids := make([]string, 0)
	query := "SELECT id FROM users " + condition
	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		if err == sql.ErrNoRows {
			return ids, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		if id != constant.UUID_USER_SYSTEM {
			ids = append(ids, id)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func GetAccount(ctx context.Context, logger runtime.Logger, db *sql.DB, userID string) (*entity.Account, error) {
	var displayName sql.NullString
	var username sql.NullString
	var avatarURL sql.NullString
	var langTag sql.NullString
	var location sql.NullString
	var timezone sql.NullString
	var metadata sql.NullString
	var wallet sql.NullString
	var email sql.NullString
	var apple sql.NullString
	var facebook sql.NullString
	var facebookInstantGame sql.NullString
	var google sql.NullString
	var gamecenter sql.NullString
	var steam sql.NullString
	var customID sql.NullString
	var edgeCount int
	var createTime pgtype.Timestamptz
	var updateTime pgtype.Timestamptz
	var verifyTime pgtype.Timestamptz
	var disableTime pgtype.Timestamptz
	var deviceIDs pgtype.VarcharArray
	var lastOnlineTime pgtype.Timestamptz

	query := `
SELECT u.username, u.display_name, u.avatar_url, u.lang_tag, u.location, u.timezone, u.metadata, u.wallet,
	u.email, u.apple_id, u.facebook_id, u.facebook_instant_game_id, u.google_id, u.gamecenter_id, u.steam_id, u.custom_id, u.edge_count,
	u.create_time, u.update_time, u.verify_time, u.disable_time, array(select ud.id from user_device ud where u.id = ud.user_id), u.last_online_time_unix
FROM users u
WHERE u.id = $1`

	if err := db.QueryRowContext(ctx, query, userID).
		Scan(&username, &displayName, &avatarURL,
			&langTag, &location, &timezone,
			&metadata, &wallet, &email,
			&apple, &facebook, &facebookInstantGame,
			&google, &gamecenter, &steam,
			&customID, &edgeCount, &createTime,
			&updateTime, &verifyTime, &disableTime, &deviceIDs, &lastOnlineTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAccountNotFound
		}
		logger.Error("Error retrieving user account.%s", err.Error())
		return nil, err
	}

	devices := make([]*api.AccountDevice, 0, len(deviceIDs.Elements))
	for _, deviceID := range deviceIDs.Elements {
		devices = append(devices, &api.AccountDevice{Id: deviceID.String})
	}

	var verifyTimestamp *timestamppb.Timestamp
	if verifyTime.Status == pgtype.Present && verifyTime.Time.Unix() != 0 {
		verifyTimestamp = &timestamppb.Timestamp{Seconds: verifyTime.Time.Unix()}
	}
	var disableTimestamp *timestamppb.Timestamp
	if disableTime.Status == pgtype.Present && disableTime.Time.Unix() != 0 {
		disableTimestamp = &timestamppb.Timestamp{Seconds: disableTime.Time.Unix()}
	}
	account := entity.Account{
		Account: api.Account{
			User: &api.User{
				Id:                    userID,
				Username:              username.String,
				DisplayName:           displayName.String,
				AvatarUrl:             avatarURL.String,
				LangTag:               langTag.String,
				Location:              location.String,
				Timezone:              timezone.String,
				Metadata:              metadata.String,
				AppleId:               apple.String,
				FacebookId:            facebook.String,
				FacebookInstantGameId: facebookInstantGame.String,
				GoogleId:              google.String,
				GamecenterId:          gamecenter.String,
				SteamId:               steam.String,
				EdgeCount:             int32(edgeCount),
				CreateTime:            &timestamppb.Timestamp{Seconds: createTime.Time.Unix()},
				UpdateTime:            &timestamppb.Timestamp{Seconds: updateTime.Time.Unix()},
			},
			Wallet:      wallet.String,
			Email:       email.String,
			Devices:     devices,
			CustomId:    customID.String,
			VerifyTime:  verifyTimestamp,
			DisableTime: disableTimestamp,
		},
		LastOnlineTimeUnix: lastOnlineTime.Time.Unix(),
	}
	return &account, nil
}
