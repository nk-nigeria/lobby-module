package api

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ciaolink-game-platform/cgb-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgb-lobby-module/cgbdb"
	"github.com/ciaolink-game-platform/cgb-lobby-module/conf"
	pb "github.com/ciaolink-game-platform/cgp-common/proto"
	"github.com/heroiclabs/nakama-common/runtime"
)

func RpcGetAllExchange() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		// if userID != "" {
		// 	return "", errors.New("Unath.")
		// }
		unmarshaler := conf.Unmarshaler
		exChangedealReq := &pb.ExchangeRequest{}

		if payload != "" {
			if err := unmarshaler.Unmarshal([]byte(payload), exChangedealReq); err != nil {
				logger.Error("Error when unmarshal payload", err.Error())
				return "", presenter.ErrUnmarshal
			}
		}
		if userID != "" {
			exChangedealReq.UserIdRequest = userID
		}
		list, err := cgbdb.GetAllExchange(ctx, logger, db, userID, exChangedealReq)
		if err != nil {
			logger.Error("Error when get all list exchange, err %s", err.Error())
			return "", presenter.ErrUnmarshal
		}
		sarshaler := conf.MarshalerDefault
		listJson, _ := sarshaler.Marshal(list)
		logger.Info(string(listJson))
		return string(listJson), nil
	}
}

func RpcExchangeLock() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unath.")
		}
		unmarshaler := conf.Unmarshaler
		exChangedealReq := &pb.ExchangeInfo{}
		if err := unmarshaler.Unmarshal([]byte(payload), exChangedealReq); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		exchangeDB, err := cgbdb.ExchangeLock(ctx, logger, db, exChangedealReq)
		if err != nil {
			logger.Error("Error when lock exchange  %s, err %s", exChangedealReq.GetId(), err.Error())
			return "", presenter.ErrUnmarshal
		}
		sarshaler := conf.MarshalerDefault
		listJson, _ := sarshaler.Marshal(exchangeDB)
		return string(listJson), nil
	}
}

func RpcGetExchange() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unath.")
		}
		unmarshaler := conf.Unmarshaler
		exChangedealReq := &pb.ExchangeInfo{}
		if err := unmarshaler.Unmarshal([]byte(payload), exChangedealReq); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		exchangeDB, err := cgbdb.GetExchangeById(ctx, logger, db, exChangedealReq)
		if err != nil {
			logger.Error("Error when lock exchange  %s, err %s", exChangedealReq.GetId(), err.Error())
			return "", presenter.ErrUnmarshal
		}
		sarshaler := conf.MarshalerDefault
		strJson, _ := sarshaler.Marshal(exchangeDB)
		return string(strJson), nil
	}
}

func RpcExchangeUpdateStatus() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unath.")
		}
		unmarshaler := conf.Unmarshaler
		exChangedealReq := &pb.ExchangeInfo{}
		if err := unmarshaler.Unmarshal([]byte(payload), exChangedealReq); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		exchangeDB, err := cgbdb.ExchangeUpdateStatus(ctx, logger, db, exChangedealReq)
		if err != nil {
			logger.Error("Error when update status exchange  %s, err %s", exChangedealReq.GetId(), err.Error())
			return "", presenter.ErrUnmarshal
		}
		sarshaler := conf.MarshalerDefault
		strJson, _ := sarshaler.Marshal(exchangeDB)
		return string(strJson), nil
	}
}
