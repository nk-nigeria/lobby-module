package api

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	pb "github.com/nakama-nigeria/cgp-common/proto"
	"github.com/nakama-nigeria/cgp-common/utilities"
	"github.com/nakama-nigeria/lobby-module/api/presenter"
	"github.com/nakama-nigeria/lobby-module/cgbdb"
	"google.golang.org/protobuf/proto"
)

func RpcUserChangePass(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {

		userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		request := &pb.ChangePasswordRequest{}
		if err := utilities.DecodeBase64Proto(payload, request); err != nil {
			return "", err
		}

		err := cgbdb.ChangePasswordUser(ctx, logger, db,
			userId, request.GetOldPassword(), request.GetPassword())
		if err != nil {
			logger.Error("Change password user %s, error: %s", userId, err.Error())
			return "", err
		}
		return "", nil
	}
}

func RpcLinkUsername(marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions) func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, _ string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		logger.Info("request link username")
		// parse request
		userId, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			logger.Error("request link username userid error", ok)
			return "", presenter.ErrNoUserIdFound
		}

		request := &pb.RegisterRequest{}
		if err := utilities.DecodeBase64Proto(payload, request); err != nil {
			return "", err
		}

		logger.Info("user %s request register %v", userId, request)
		err := cgbdb.LinkUsername(ctx, logger, db, userId, request.UserName, request.Password)
		if err != nil {
			logger.Error("link username error", err)
		}

		return "", err
	}
}

func AfterAuthDevice(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out *api.Session, in *api.AuthenticateDeviceRequest) error {
	logger.Debug("AfterAuthDevice: %s", out.Token)

	userID, err := getUserIDFromToken(out.Token)
	if err != nil {
		logger.Error("Failed to parse user ID from token: %v", err)
		return err
	}

	var exists bool
	err = db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM users_ext WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		sid := createSidFromUserID(userID)
		_, err := db.ExecContext(ctx, "INSERT INTO users_ext (id, sid) VALUES ($1, $2)", userID, sid)
		if err != nil {
			logger.Error("Insert users_ext failed: %v", err)
			return err
		}
		logger.Debug("Inserted users_ext for user %s", userID)
	}

	return nil
}

var jwtSecret = []byte("YM0fahhp2aW9dcuqq2Z6gFj6AvyaJLAfMMofaCmJ91c=") // phải giống với giá trị `session.token_signing_key` trong nakama.yml

func getUserIDFromToken(tokenStr string) (string, error) {
	tokenStr = strings.TrimSpace(tokenStr)
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Nakama dùng HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uid, ok := claims["uid"].(string)
		if !ok {
			return "", jwt.ErrInvalidKey
		}
		return uid, nil
	}
	return "", jwt.ErrInvalidKey
}

func createSidFromUserID(userID string) int64 {
	hash := sha256.Sum256([]byte(userID))
	value := binary.BigEndian.Uint64(hash[:8])
	sid := value % 1_000_000_000

	if sid < 100_000_000 {
		sid += 100_000_000
	}
	return int64(sid)
}
