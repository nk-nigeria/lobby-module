package cgbdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	"github.com/nakamaFramework/cgb-lobby-module/constant"
	"github.com/nakamaFramework/cgb-lobby-module/entity"
	objectstorage "github.com/nakamaFramework/cgb-lobby-module/object-storage"
	pb "github.com/nakamaFramework/cgp-common/proto"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const DefaultLevel = 0

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
func GetAccount(ctx context.Context, db *sql.DB, userID string, userSid int64) (*entity.Account, error) {
	if len(userID) == 0 && userSid <= 0 {
		return nil, ErrAccountNotFound
	}

	query := `
SELECT u.id,u.username, u.display_name, u.avatar_url, u.lang_tag, u.location, u.timezone, u.metadata, u.wallet,
	u.email, u.apple_id, u.facebook_id, u.facebook_instant_game_id, u.google_id, u.gamecenter_id, u.steam_id, u.custom_id, u.edge_count,
	u.create_time, u.update_time, u.verify_time, u.disable_time, array(select ud.id from user_device ud where u.id = ud.user_id),
	u.sid
FROM users u
WHERE `
	args := make([]any, 0)
	if len(userID) > 0 {
		query += ` u.id = $1`
		args = append(args, userID)
	} else if userSid > 0 {
		query += ` u.sid = $1`
		args = append(args, userSid)
	}

	row := db.QueryRowContext(ctx, query, args...)
	account, err := scanAccount(row)
	if err == sql.ErrNoRows {
		return nil, ErrAccountNotFound
	}
	return account, err
}

func GetAccounts(ctx context.Context, db *sql.DB, userIds ...string) ([]*entity.Account, error) {
	if len(userIds) == 0 {
		return make([]*entity.Account, 0), nil
	}
	query := `
SELECT u.id, u.username, u.display_name, u.avatar_url, u.lang_tag, u.location, u.timezone, u.metadata, u.wallet,
	u.email, u.apple_id, u.facebook_id, u.facebook_instant_game_id, u.google_id, u.gamecenter_id, u.steam_id, u.custom_id, u.edge_count,
	u.create_time, u.update_time, u.verify_time, u.disable_time, array(select ud.id from user_device ud where u.id = ud.user_id),
	u.sid
FROM users u
WHERE u.id::text IN (` + "'" + strings.Join(userIds, "','") + "'" + `)`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	ml := make([]*entity.Account, 0)
	for rows.Next() {
		account, err := scanAccount(rows)
		if err != nil {
			continue
		}
		ml = append(ml, account)
	}
	return ml, nil
}

func GetProfileUser(ctx context.Context, db *sql.DB, userID string, objStorage objectstorage.ObjStorage) (*pb.Profile, map[string]interface{}, error) {
	// account, err := nk.AccountGetId(ctx, userID)
	account, err := GetAccount(ctx, db, userID, 0)
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
		CreateTimeUnix:     user.GetCreateTime().Seconds,
		// LangAvailables:     []string{"en", "phi"},
	}
	playingMatchJson := entity.InterfaceToString(metadata["playing_in_match"])
	profile.PlayingMatch = &pb.PlayingMatch{}
	if len(playingMatchJson) > 0 {
		conf.Unmarshaler.Unmarshal([]byte(playingMatchJson), profile.PlayingMatch)
	}

	profile.LangAvailables = append(profile.LangAvailables,
		&pb.LangCode{
			IsoCode:     "en-US",
			DisplayName: "English",
		},
		&pb.LangCode{
			IsoCode:     "tl-PH",
			DisplayName: "Philippines",
		},
	)
	if objStorage != nil {
		for _, s := range profile.LangAvailables {
			sourceUrl, _ := objStorage.PresignGetObject("lang", s.IsoCode+".json", 24*time.Hour, nil)
			s.SourceUrl = strings.Split(sourceUrl, ".json")[0] + ".json"
		}
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

	if profile.RefCode == "" {
		profile.RemainTimeInputRefCode = entity.MaxIn64(int64(time.Until(time.Unix(profile.CreateTimeUnix+7*86400, 0)).Seconds()), 0)
	}
	profile.UserSid = account.Sid
	return &profile, metadata, nil
}

type DBScan interface {
	Scan(dest ...any) error
}

func scanAccount(row DBScan) (*entity.Account, error) {
	var userId sql.NullString
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
	// var lastOnlineTime pgtype.Timestamptz
	var sID sql.NullInt64
	err := row.Scan(&userId, &username, &displayName, &avatarURL,
		&langTag, &location, &timezone,
		&metadata, &wallet, &email,
		&apple, &facebook, &facebookInstantGame,
		&google, &gamecenter, &steam,
		&customID, &edgeCount, &createTime,
		&updateTime, &verifyTime, &disableTime,
		&deviceIDs, &sID)
	if err != nil {
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
	account := &entity.Account{
		Account: api.Account{
			User: &api.User{
				Id:                    userId.String,
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
		Sid: sID.Int64,
	}
	return account, nil
}

func UpdateUsersPlayingInMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, userId string, pl *pb.PlayingMatch) error {
	if pl == nil {
		return errors.New("invalid param")
	}
	if len(userId) == 0 {
		return nil
	}
	v := &pb.PlayingMatch{
		Code:      pl.Code,
		MatchId:   pl.MatchId,
		LeaveTime: pl.LeaveTime,
		Mcb:       pl.Mcb,
		Bet:       pl.Bet,
	}
	data, err := conf.Marshaler.Marshal(v)
	if err != nil {
		return err
	}
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(
		`UPDATE
					users AS u
				SET
					metadata
						= u.metadata
						|| jsonb_build_object('playing_in_match','` + string(data) + `' )
				WHERE	
					id =$1`)
	query := queryBuilder.String()
	_, err = db.ExecContext(ctx, query, userId)
	if err != nil {
		logger.WithField("err", err).Error("db.ExecContext match update error.")
	}
	return err
}

type ListProfile []*pb.SimpleProfile

func (l ListProfile) ToMap() map[string]*pb.SimpleProfile {
	mapProfile := make(map[string]*pb.SimpleProfile)
	for _, p := range l {
		mapProfile[p.GetUserId()] = p
	}
	return mapProfile
}
func GetProfileUsers(ctx context.Context, db *sql.DB, userIDs ...string) (ListProfile, error) {
	// accounts, err := nk.AccountsGetId(ctx, userIDs)
	accounts, err := GetAccounts(ctx, db, userIDs...)
	if err != nil {
		return nil, err
	}
	listProfile := make(ListProfile, 0, len(accounts))
	for _, acc := range accounts {
		u := acc.GetUser()
		var metadata map[string]interface{}
		json.Unmarshal([]byte(u.GetMetadata()), &metadata)
		profile := pb.SimpleProfile{
			UserId:      u.GetId(),
			UserName:    u.GetUsername(),
			DisplayName: u.GetDisplayName(),
			Status:      entity.InterfaceToString(metadata["status"]),
			AvatarId:    entity.InterfaceToString(metadata["avatar_id"]),
			VipLevel:    entity.ToInt64(metadata["vip_level"], 0),
			UserSid:     acc.Sid,
		}
		playingMatchJson := entity.InterfaceToString(metadata["playing_in_match"])

		if playingMatchJson == "" {
			profile.PlayingMatch = nil
		} else {
			profile.PlayingMatch = &pb.PlayingMatch{}
			json.Unmarshal([]byte(playingMatchJson), profile.PlayingMatch)
		}
		if acc.GetWallet() != "" {
			wallet, err := entity.ParseWallet(acc.GetWallet())
			if err == nil {
				profile.AccountChip = wallet.Chips
			}
		}
		listProfile = append(listProfile, &profile)
	}
	return listProfile, nil
}

func CreateNewUser(ctx context.Context, db *sql.DB, user *api.Account) error {
	stmt, err := db.Prepare("INSERT INTO public.users (id, username, display_name, avatar_url, lang_tag, location, timezone, metadata, wallet, email, password, edge_count, create_time, update_time, verify_time, disable_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	var createTime, updateTime, verifyTime, disableTime time.Time
	if user.User.CreateTime != nil {
		createTime = user.User.CreateTime.AsTime()
	}
	if user.User.UpdateTime != nil {
		updateTime = user.User.UpdateTime.AsTime()
	}
	if user.VerifyTime != nil {
		verifyTime = user.VerifyTime.AsTime()
	}
	if user.DisableTime != nil {
		disableTime = user.DisableTime.AsTime()
	}
	// Execute the SQL statement with appropriate values for the new row
	_, err = stmt.Exec(
		user.User.Id,
		user.User.Username,
		user.User.DisplayName,
		user.User.AvatarUrl,
		"en",
		nil,
		nil,
		user.User.Metadata, // JSONB metadata
		"{}",               // JSONB wallet
		user.Email,
		uuid.New().String(), // bytea password
		0,                   // edge_count
		createTime,          // create_time
		updateTime,          // update_time
		verifyTime,          // verify_time
		disableTime,         // disable_time
	)
	if err != nil {
		return err
	}
	return nil
}
