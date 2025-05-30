package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	"github.com/nakamaFramework/cgb-lobby-module/cgbdb"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	"github.com/nakamaFramework/cgp-common/lib"
	pb "github.com/nakamaFramework/cgp-common/proto"
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
		// logger.Info(string(listJson))
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
		if exchangeDB.GetStatus() == int64(pb.ExchangeStatus_EXCHANGE_STATUS_DONE.Number()) {
			props := make(map[string]string)
			props["user_id"] = exchangeDB.GetUserIdRequest()
			// TODO: fix currency_unit_id
			props["currency_unit_id"] = "1"
			props["currency_value"] = exChangedealReq.Price
			// TODO: fix publisher
			props["publisher"] = "1"
			props["time_unix"] = strconv.FormatInt(time.Now().Unix(), 10)
			props["chips"] = strconv.FormatInt(exchangeDB.Chips, 10)
			props["trans_id"] = exchangeDB.Id
			data, _ := json.Marshal(props)
			op := lib.NewReportGame(ctx)
			data, status, err := op.ReportEvent(ctx, "cashout", exchangeDB.GetUserIdRequest(), string(data))
			if err != nil || status > 300 {
				logger.Error("Report cashout %s -> %s url failed, response %s status %d err %v",
					userID, exchangeDB.Id, string(data), status, err)

			} else {
				logger.Info("Report iap %s -> %s successful, data %s", userID, exchangeDB.Id, string(data))
			}
		}
		return string(strJson), nil
	}
}
