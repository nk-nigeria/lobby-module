package cgbdb

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgconn"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func AddNewUser(ctx context.Context, logger runtime.Logger, db *sql.DB, username, password, customid string) (string, error) {
	userID := uuid.Must(uuid.NewV4()).String()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Error hashing password username %s, err: %s", username, password)
		return "", status.Error(codes.Internal, "Error finding or creating user account.")
	}
	query := "INSERT INTO users (id, username, password, custom_id, create_time, update_time) VALUES ($1, $2, $3, $4, now(), now())"
	result, err := db.ExecContext(ctx, query, userID, username, hashedPassword, customid)
	if err != nil {
		logger.Error("Query error %s", err.Error())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == DbErrorUniqueViolation {
			if strings.Contains(pgErr.Message, "users_username_key") {
				// Username is already in use by a different account.
				return "", status.Error(codes.AlreadyExists, "Username is already in use.")
			}
		}
		logger.Error("Cannot find or create user with username: %s, err: %s", username, err.Error())
		return "", status.Error(codes.Internal, "Error finding or creating user account.")
	}

	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert new user.")
		return "", status.Error(codes.Internal, "Error finding or creating user account.")
	}
	return userID, nil
}

func ChangePasswordUser(ctx context.Context, logger runtime.Logger, db *sql.DB, username, oldpassword, newpassword, customid string) error {
	hashedOldPassword, err := bcrypt.GenerateFromPassword([]byte(oldpassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Error hashing old password username %s, err: %s", username, err.Error())
		return status.Error(codes.Internal, "Error finding or creating user account.")
	}
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newpassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Error hashing new password username %s, err: %s", username, err.Error())
	}
	query := "UPDATE users SET password=$1 WHERE username=$2 AND $password=$3"

	result, err := db.ExecContext(ctx, query, hashedNewPassword, username, hashedOldPassword)
	if err != nil {
		logger.Error("Cannot update password user with username %s, err: %s", username, err.Error())
		return status.Error(codes.Internal, "Error update password user account.")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert new user.")
		return status.Error(codes.Internal, "Error finding user account.")
	}
	return nil
}
