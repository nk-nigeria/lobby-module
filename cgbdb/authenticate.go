package cgbdb

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func AuthenticateCustomUsername(ctx context.Context, logger runtime.Logger, db *sql.DB, customid, username, password string, create bool) (*entity.CustomUser, bool, error) {
	logger.Info("AuthenticateUsername %s", username)
	customUser := &entity.CustomUser{
		Id:       customid,
		UserName: username,
	}
	// Look for an existing account.
	query := "SELECT id, password, custom_id, disable_time FROM users WHERE username = $1"
	var dbUserID string
	var dbPassword []byte
	var dbCustomId string
	var dbDisableTime pgtype.Timestamptz
	found := true

	err := db.QueryRowContext(ctx, query, username).Scan(&dbUserID, &dbPassword, &dbCustomId, &dbDisableTime)
	if err != nil {
		if err == sql.ErrNoRows {
			// Account not found and creation is never allowed for this type.
			// return "", status.Error(codes.NotFound, "User account not found.")
			found = false
		} else {
			logger.Error("Error looking up user by username.", zap.Error(err), zap.String("username", username))
			return nil, false, status.Error(codes.Internal, "Error finding user account.")
		}
	}
	if found {
		// Check if it's disabled.
		if dbDisableTime.Status == pgtype.Present && dbDisableTime.Time.Unix() != 0 {
			logger.Info("User account is disabled.", zap.String("username", username))
			return nil, false, status.Error(codes.PermissionDenied, "User account banned.")
		}

		// Check if the account has a password.
		if len(dbPassword) == 0 {
			// Do not disambiguate between bad password and password login not possible at all in client-facing error messages.
			return nil, false, status.Error(codes.Unauthenticated, "Invalid credentials.")
		}

		// Check if password matches.
		err = bcrypt.CompareHashAndPassword(dbPassword, []byte(password))
		if err != nil {
			return nil, false, status.Error(codes.Unauthenticated, "Invalid credentials.")
		}
		customUser.Id = dbCustomId
		customUser.UserId = dbUserID
		return customUser, false, nil
	}
	if !create {
		// No user account found, and creation is not allowed.
		return nil, false, status.Error(codes.NotFound, "User account not found.")
	}
	newUserId, err := AddNewUser(ctx, logger, db, username, password, customid)
	if err != nil {
		return nil, false, err
	}
	customUser.UserId = newUserId
	return customUser, true, nil
}

func LinkDeviceWithNewUser(ctx context.Context, logger runtime.Logger, db *sql.DB, deviceId, customid, username, password string, create bool) (*entity.CustomUser, bool, error) {

	if deviceId == "" {
		logger.Info("deviceId empty, call AuthenticateCustomUsername %s", username)
		return AuthenticateCustomUsername(ctx, logger, db, customid, username, password, create)
	}
	logger.Info("LinkDeviceWithNewUser %s", username)
	customUser := &entity.CustomUser{
		Id:       customid,
		UserName: username,
	}
	query := "SELECT user_id FROM user_device WHERE id = $1"
	var dbUserID string

	err := db.QueryRowContext(ctx, query, deviceId).Scan(&dbUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Info("deviceId %s not found, call AuthenticateCustomUsername %s", deviceId, username)
			// if not found device id, create new user
			return AuthenticateCustomUsername(ctx, logger, db, customid, username, password, create)
		} else {
			logger.Error("Error looking up user_id by deviceId. %s, err: %s", deviceId, err.Error())
			return nil, false, status.Error(codes.Internal, "Error finding user account.")
		}
	}
	// found user_id link with device id, update username, password if password is nil
	query = "SELECT id, username, password, custom_id, disable_time FROM users WHERE id = $1"
	var dbPassword []byte
	var dbUsername string
	var dbCustomId string
	var dbDisableTime pgtype.Timestamptz

	err = db.QueryRowContext(ctx, query, dbUserID).Scan(&dbUserID, &dbUsername, &dbPassword, &dbCustomId, &dbDisableTime)
	if err != nil {
		logger.Error("Error looking up user by username %s, err: %s", username, err.Error())
		return nil, false, status.Error(codes.Internal, "Error finding user account.")
	}
	// Check if it's disabled.
	if dbDisableTime.Status == pgtype.Present && dbDisableTime.Time.Unix() != 0 {
		logger.Info("User account is disabled, username %s", username)
		return nil, false, status.Error(codes.PermissionDenied, "User account banned.")
	}

	// Check if the account has a password.
	if len(dbPassword) == 0 {
		logger.Info("Update user username %s, Link device %s ", username, deviceId)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("Error hashing password. username %s", username)
			return nil, false, status.Error(codes.Internal, "Error finding or creating user account.")
		}
		// update user
		query = "UPDATE users SET username=$1, password=$2, custom_id=$3 WHERE id=$4"
		_, err = db.ExecContext(ctx, query, username, hashedPassword, customid, dbUserID)
		if err != nil {
			logger.Error("Cannot update user with username %s, err: %s", username, err.Error())
			return nil, false, status.Error(codes.Internal, "Error finding or creating user account.")
		}
		return AuthenticateCustomUsername(ctx, logger, db, customid, username, password, false)
	} else {
		if username != dbUsername || bcrypt.CompareHashAndPassword(dbPassword, []byte(password)) != nil {
			return AuthenticateCustomUsername(ctx, logger, db, customid, username, password, false)
		}
	}

	// Check if password matches.
	customUser.UserId = dbUserID
	customUser.Id = dbCustomId
	return customUser, false, nil
}

func AuthenticateEmail(ctx context.Context, logger runtime.Logger, db *sql.DB, email, password, username string, create bool) (string, string, bool, error) {
	found := true

	// Look for an existing account.
	query := "SELECT id, username, password, disable_time FROM users WHERE email = $1"
	var dbUserID string
	var dbUsername string
	var dbPassword []byte
	var dbDisableTime pgtype.Timestamptz
	err := db.QueryRowContext(ctx, query, email).Scan(&dbUserID, &dbUsername, &dbPassword, &dbDisableTime)
	if err != nil {
		if err == sql.ErrNoRows {
			found = false
		} else {
			logger.Error("Error looking up user by email.", zap.Error(err), zap.String("email", email), zap.String("username", username), zap.Bool("create", create))
			return "", "", false, status.Error(codes.Internal, "Error finding user account.")
		}
	}

	// Existing account found.
	if found {
		// Check if it's disabled.
		if dbDisableTime.Status == pgtype.Present && dbDisableTime.Time.Unix() != 0 {
			logger.Info("User account is disabled.", zap.String("email", email), zap.String("username", username), zap.Bool("create", create))
			return "", "", false, status.Error(codes.PermissionDenied, "User account banned.")
		}

		// Check if password matches.
		err = bcrypt.CompareHashAndPassword(dbPassword, []byte(password))
		if err != nil {
			return "", "", false, status.Error(codes.Unauthenticated, "Invalid credentials.")
		}

		return dbUserID, dbUsername, false, nil
	}

	if !create {
		// No user account found, and creation is not allowed.
		return "", "", false, status.Error(codes.NotFound, "User account not found.")
	}

	// Create a new account.
	userID := uuid.Must(uuid.NewV4()).String()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Error hashing password.", zap.Error(err), zap.String("email", email), zap.String("username", username), zap.Bool("create", create))
		return "", "", false, status.Error(codes.Internal, "Error finding or creating user account.")
	}
	query = "INSERT INTO users (id, username, email, password, create_time, update_time) VALUES ($1, $2, $3, $4, now(), now())"
	result, err := db.ExecContext(ctx, query, userID, username, email, hashedPassword)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == DbErrorUniqueViolation {
			if strings.Contains(pgErr.Message, "users_username_key") {
				// Username is already in use by a different account.
				return "", "", false, status.Error(codes.AlreadyExists, "Username is already in use.")
			} else if strings.Contains(pgErr.Message, "users_email_key") {
				// A concurrent write has inserted this email.
				logger.Info("Did not insert new user as email already exists.", zap.Error(err), zap.String("email", email), zap.String("username", username), zap.Bool("create", create))
				return "", "", false, status.Error(codes.Internal, "Error finding or creating user account.")
			}
		}
		logger.Error("Cannot find or create user with email.", zap.Error(err), zap.String("email", email), zap.String("username", username), zap.Bool("create", create))
		return "", "", false, status.Error(codes.Internal, "Error finding or creating user account.")
	}

	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert new user.", zap.Int64("rows_affected", rowsAffectedCount))
		return "", "", false, status.Error(codes.Internal, "Error finding or creating user account.")
	}

	return userID, username, true, nil
}
