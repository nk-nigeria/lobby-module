package cgbdb

import (
	"context"
	"database/sql"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"regexp"
	"strings"
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
		return status.Error(codes.InvalidArgument, "Username address and password is required.")
	} else if invalidUsernameRegex.MatchString(username) {
		return status.Error(codes.InvalidArgument, "Invalid username, no spaces or control characters allowed.")
	} else if len(password) < 8 {
		return status.Error(codes.InvalidArgument, "Password must be at least 8 characters long.")
	} else if len(username) < 8 || len(username) > 255 {
		return status.Error(codes.InvalidArgument, "Invalid username address, must be 10-255 bytes.")
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
		return status.Error(codes.AlreadyExists, "Username is already in use.")
	}
	return nil
}
