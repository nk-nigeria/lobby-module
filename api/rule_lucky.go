package api

import (
	"context"
	"database/sql"
	"errors"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	"github.com/nakamaFramework/cgb-lobby-module/cgbdb"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	"github.com/nakamaFramework/cgp-common/define"
	pb "github.com/nakamaFramework/cgp-common/proto"
)

func RpcRuleLucky() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unath.")
		}
		unmarshaler := conf.Unmarshaler
		req := &pb.RuleLucky{}
		if len(payload) > 0 {
			err := unmarshaler.Unmarshal([]byte(payload), req)
			if err != nil {
				logger.WithField("err", err).Error("Error when unmarshal payload")
				return "", presenter.ErrUnmarshal
			}
		}
		ml, err := cgbdb.QueryRulesLucky(ctx, db, req)
		if err != nil {
			logger.WithField("err", err).Error("Error when query rules lucky")
			return "", presenter.ErrInternalError
		}
		dataJson, _ := conf.MarshalerDefault.Marshal(&pb.RulesLucky{
			Rules: ml,
			Total: int64(len(ml)),
		})
		return string(dataJson), nil
	}
}

func RpcRuleLuckyAdd() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unath.")
		}
		if len(payload) == 0 {
			logger.Error("Error when unmarshal is empty payload")
			return "", presenter.ErrInvalidInput
		}
		unmarshaler := conf.Unmarshaler
		req := &pb.RuleLucky{}
		err := unmarshaler.Unmarshal([]byte(payload), req)
		if err != nil {
			logger.WithField("err", err).Error("Error when unmarshal payload")
			return "", presenter.ErrUnmarshal
		}
		err = cgbdb.InsertRulesLucky(ctx, db, req)
		if err != nil {
			logger.WithField("err", err).Error("Error when insert rules lucky")
			return "", presenter.ErrInternalError
		}
		dataJson, _ := conf.MarshalerDefault.Marshal(req)
		// nk.Event(ctx, &api.Event{
		// 	Name:       define.NakEventRuleLuckyChange,
		// 	Properties: map[string]string{"data": string(dataJson)},
		// })
		return string(dataJson), nil
	}
}

func RpcRuleLuckyUpdate() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unath.")
		}
		if len(payload) == 0 {
			logger.Error("Error when unmarshal is empty payload")
			return "", presenter.ErrInvalidInput
		}
		unmarshaler := conf.Unmarshaler
		req := &pb.RuleLucky{}
		err := unmarshaler.Unmarshal([]byte(payload), req)
		if err != nil {
			logger.WithField("err", err).Error("Error when unmarshal payload")
			return "", presenter.ErrUnmarshal
		}
		result, err := cgbdb.UpdateRulesLucky(ctx, db, req)
		if err != nil {
			logger.WithField("err", err).Error("Error when insert rules lucky")
			return "", presenter.ErrInternalError
		}
		dataJson, _ := conf.MarshalerDefault.Marshal(result)

		return string(dataJson), nil
	}
}

func RpcRuleLuckyDelete() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unath.")
		}
		if len(payload) == 0 {
			logger.Error("Error when unmarshal is empty payload")
			return "", presenter.ErrInvalidInput
		}
		unmarshaler := conf.Unmarshaler
		req := &pb.RuleLucky{}
		err := unmarshaler.Unmarshal([]byte(payload), req)
		if err != nil {
			logger.WithField("err", err).Error("Error when unmarshal payload")
			return "", presenter.ErrUnmarshal
		}
		if req.Id <= 0 {
			return "", nil
		}
		err = cgbdb.DeleteRulesLucky(ctx, db, req.Id)
		if err != nil {
			logger.WithField("err", err).Error("Error when insert rules lucky")
			return "", presenter.ErrInternalError
		}
		return "", nil
	}
}

func RpcRuleLuckyEmitEvemt() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unath.")
		}
		if len(payload) == 0 {
			logger.Error("Error when unmarshal is empty payload")
			return "", presenter.ErrInvalidInput
		}
		unmarshaler := conf.Unmarshaler
		req := &pb.RuleLucky{}
		err := unmarshaler.Unmarshal([]byte(payload), req)
		if err != nil {
			logger.WithField("err", err).Error("Error when unmarshal payload")
			return "", presenter.ErrUnmarshal
		}
		if len(req.GameCode) == 0 {
			return "", nil
		}
		_ = cgbdb.UpdateEmitEventLucky(ctx, db, req)
		nk.Event(ctx, &api.Event{
			Name:       define.NakEventRuleLuckyChange,
			Properties: map[string]string{"game_code": req.GameCode},
		})
		return "", nil

	}
}
