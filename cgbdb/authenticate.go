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
	query := "SELECT id, password, disable_time FROM users WHERE username = $1"
	var dbUserID string
	var dbPassword []byte
	var dbDisableTime pgtype.Timestamptz
	found := true

	err := db.QueryRowContext(ctx, query, username).Scan(&dbUserID, &dbPassword, &dbDisableTime)
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
		customUser.UserId = dbUserID
		return customUser, false, nil
	}
	if !create {
		// No user account found, and creation is not allowed.
		return nil, false, status.Error(codes.NotFound, "User account not found.")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Error hashing password.", zap.Error(err), zap.String("username", username), zap.String("username", username), zap.Bool("create", create))
		return nil, false, status.Error(codes.Internal, "Error finding or creating user account.")
	}
	// Create a new account.
	userID := uuid.Must(uuid.NewV4()).String()
	customUser.UserId = userID
	query = "INSERT INTO users (id, username, password, create_time, update_time) VALUES ($1, $2, $3, now(), now())"
	result, err := db.ExecContext(ctx, query, userID, username, hashedPassword)
	if err != nil {
		logger.Error("Query error %s", err.Error())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == DbErrorUniqueViolation {
			if strings.Contains(pgErr.Message, "users_username_key") {
				// Username is already in use by a different account.
				return nil, false, status.Error(codes.AlreadyExists, "Username is already in use.")
			}
		}
		logger.Error("Cannot find or create user with username.", zap.Error(err), zap.String("username", username), zap.Bool("create", create))
		return nil, false, status.Error(codes.Internal, "Error finding or creating user account.")
	}

	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert new user.", zap.Int64("rows_affected", rowsAffectedCount))
		return nil, false, status.Error(codes.Internal, "Error finding or creating user account.")
	}
	return customUser, true, nil
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
