package api

import (
	"context"
	"database/sql"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

func RpcPushToBank(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		bank := &pb.Bank{}
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		if payload == "" {
			return "", presenter.ErrMarshal
		}
		unmarshaler.Unmarshal([]byte(payload), bank)
		bank.SenderId = userID
		newBank, err := entity.BankPushToSafe(ctx, logger, nk, unmarshaler, bank)
		if err != nil {
			return "", err
		}
		newBankJson, _ := marshaler.Marshal(newBank)
		return string(newBankJson), nil
	}
}

func RpcWithDraw(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		bank := &pb.Bank{}
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		if payload == "" {
			return "", presenter.ErrMarshal
		}
		unmarshaler.Unmarshal([]byte(payload), bank)
		bank.SenderId = userID
		newBank, err := entity.BankWithdraw(ctx, logger, nk, bank)
		if err != nil {
			return "", err
		}
		newBankJson, _ := marshaler.Marshal(newBank)
		return string(newBankJson), nil
	}
}
