package api

import (
	"context"
	"database/sql"
	"strings"

	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/gofrs/uuid"

	"github.com/ciaolink-game-platform/cgp-lobby-module/cgbdb"
	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func RpcAuthenticateEmail(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		queryParms, ok := ctx.Value(runtime.RUNTIME_CTX_QUERY_PARAMS).(map[string][]string)
		if !ok {
			queryParms = make(map[string][]string)
		}
		if len(queryParms) == 0 ||
			(len(queryParms["email"]) == 0 && len(queryParms["user_name"]) == 0) ||
			len(queryParms["password"]) == 0 {
			return "", status.Error(codes.InvalidArgument, "Email address and password is required.")
		}
		email := ""
		if v := queryParms["email"]; len(v) > 0 {
			email = v[0]
		}
		password := queryParms["password"][0]

		attemptUsernameLogin := false
		if email == "" {
			// Password was supplied, but no email. Perhaps the user is attempting to login with username/password.
			attemptUsernameLogin = true
		} else if entity.InvalidCharsRegex.MatchString(email) {
			return "", status.Error(codes.InvalidArgument, "Invalid email address, no spaces or control characters allowed.")
		} else if !entity.EmailRegex.MatchString(email) {
			return "", status.Error(codes.InvalidArgument, "Invalid email address format.")
		} else if len(email) < 10 || len(email) > 255 {
			return "", status.Error(codes.InvalidArgument, "Invalid email address, must be 10-255 bytes.")
		}

		if len(password) < entity.MIN_LENGTH_PASSWORD {
			return "", status.Errorf(codes.InvalidArgument,
				"Password must be at least %d characters long.", entity.MIN_LENGTH_PASSWORD)
		}
		username := ""
		if v := queryParms["user_name"]; len(v) > 0 {
			username = v[0]
		}
		if username == "" {
			// If no username was supplied and the email was missing.
			if attemptUsernameLogin {
				return "", status.Error(codes.InvalidArgument,
					"Username is required when email address is not supplied.")
			}
			// Email address was supplied, we are allowed to generate a username.
			username = entity.GenerateUsername()
		} else if entity.ValidUsernameRegex.MatchString(username) {
			return "", status.Error(codes.InvalidArgument, "Username invalid, no spaces or control characters allowed.")
		} else if len(username) > 128 {
			return "", status.Error(codes.InvalidArgument, "Username invalid, must be 1-128 bytes.")
		}

		var dbUserID string
		var err error
		create := false
		if v := queryParms["create"]; len(v) > 0 {
			create = entity.String2Bool(v[0])
		}

		// var customUser *entity.CustomUser
		if attemptUsernameLogin {
			// Attempting to log in with username/password. Create flag is ignored, creation is not possible here.
			_, _, err = cgbdb.AuthenticateCustomUsername(ctx, logger, db, "", username, password, create)
		} else {
			// Attempting email authentication, may or may not create.
			cleanEmail := strings.ToLower(email)
			dbUserID, username, _, err = cgbdb.AuthenticateEmail(ctx, logger, db, cleanEmail, password, username, create)
		}
		if err != nil {
			return "", err
		}
		// nk.AuthenticateApple()
		token, _, err := nk.AuthenticateTokenGenerate(dbUserID, username, 0, nil)
		if err != nil {
			return "", err
		}
		tokenAuth := pb.TokenAuth{
			Token: token,
		}
		// todo create refresh token
		tokenAuthJon, _ := marshaler.Marshal(&tokenAuth)
		// token, exp := generateToken(s.config, dbUserID, username, in.Account.Vars)
		// refreshToken, refreshExp := generateRefreshToken(s.config, dbUserID, username, in.Account.Vars)
		// s.sessionCache.Add(uuid.FromStringOrNil(dbUserID), exp, token, refreshExp, refreshToken)
		// session := &api.Session{Created: created, Token: token, RefreshToken: refreshToken}
		return string(tokenAuthJon), nil
	}
}

func BeforeAuthenticateCustom(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *api.AuthenticateCustomRequest) (*api.AuthenticateCustomRequest, error) {
	// format
	// [username]:[password]:[user_id_need_linked]
	arrAccount := strings.Split(in.Account.Id, ":")
	if len(arrAccount) < 2 {
		return nil, status.Error(codes.InvalidArgument, "Id invalid, must contains :")
	}
	username := arrAccount[0]
	password := arrAccount[1]
	linkUserId := ""
	if len(arrAccount) > 2 {
		linkUserId = arrAccount[2]
	}

	if len(password) < entity.MIN_LENGTH_PASSWORD {
		return nil, status.Errorf(codes.InvalidArgument,
			"Password must be at least %d characters long.", entity.MIN_LENGTH_PASSWORD)
	}
	if len(username) > 128 {
		return nil, status.Error(codes.InvalidArgument, "Username invalid, must be 1-128 bytes.")
	} else if !entity.ValidUsernameRegex.MatchString(username) {
		return nil, status.Error(codes.InvalidArgument, "Username invalid, only alphabet characters and number allowed.")
	}

	customid := uuid.Must(uuid.NewV4()).String()
	createIfNotExist := in.GetCreate().GetValue()
	customUser, created, err := cgbdb.AuthenticateCustomUsername(ctx, logger,
		db, customid, username, password, createIfNotExist)
	if err != nil {
		return nil, err
	}
	in.Account.Id = customUser.Id
	// err = nk.LinkDevice(ctx, customUser.UserId, "my_id_device_keyxxxx")
	if created && linkUserId != "" {
		err = nk.LinkCustom(ctx, linkUserId, customUser.Id)
		if err != nil {
			logger.Error("[Failed] Link custom id %s to userId %s, err: ", err.Error())
		} else {
			logger.Info("[Success] Link custom id %s to userId %s", customUser.Id, linkUserId)
		}
	}
	logger.Info("Return user id: %s, username: %s", in.Account.Id, in.Username)
	return in, nil
}

func AfterAuthenticateCustom(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out *api.Session, in *api.AuthenticateCustomRequest) error {
	// err := nk.LinkDevice(ctx, in.GetAccount().Id, "my_id_device_keyxxxx")
	// if err != nil {
	// 	logger.Error(err.Error())
	// }
	return nil
}
