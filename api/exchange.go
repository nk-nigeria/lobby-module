package api

import (
	"context"
	"database/sql"
	"errors"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	"github.com/nakamaFramework/cgb-lobby-module/cgbdb"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	pb "github.com/nakamaFramework/cgp-common/proto"
)

const (
	kExchangeCollection = "exchange-deal-collection"
	kExchangeKey        = "exchange-deal-key"
)

var MapExchangeDeal = make(map[string]*pb.Deal, 0)

func InitExchangeList(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) {
	objectIds := []*runtime.StorageRead{
		{
			Collection: kExchangeCollection,
			Key:        kExchangeKey,
		},
	}

	objects, err := nk.StorageRead(ctx, objectIds)
	if err != nil {
		logger.Error("Error when read exchange deal at init, error %s", err.Error())
	}

	if len(objects) > 0 {
		exChangeDealInShop := &pb.ExchangeDealInShop{}
		_ = conf.Unmarshaler.Unmarshal([]byte(objects[0].GetValue()), exChangeDealInShop)

		for _, deal := range exChangeDealInShop.Gcashes {
			MapExchangeDeal[deal.GetId()] = deal
		}
		logger.Info("List exchange deal already write in collection")
		return
	}
	exChangeDeals := pb.ExchangeDealInShop{
		Gcashes: []*pb.Deal{
			{
				Id:          "id_best_exchange_deal",
				Chips:       1000,
				Bonus:       0,
				Price:       "2000",
				AmountChips: 0,
				Name:        "Best deal",
				Currency:    "VND",
				Percent:     "15.5",
			},
			{
				Id:          "id_best_exchange_deal_2",
				Chips:       1000,
				Bonus:       0,
				Price:       "10000",
				AmountChips: 0,
				Name:        "Best deal",
				Currency:    "VND",
				Percent:     "33.3",
			},
		},
	}
	for _, deal := range exChangeDeals.Gcashes {
		MapExchangeDeal[deal.GetId()] = deal
	}
	marshaler := conf.Marshaler
	exChangedealsJson, err := marshaler.Marshal(&exChangeDeals)
	if err != nil {
		logger.Debug("Can not marshaler exchaneg deals for collection")
		return
	}

	writeObjects := []*runtime.StorageWrite{
		{
			Collection:      kExchangeCollection,
			Key:             kExchangeKey,
			Value:           string(exChangedealsJson),
			PermissionRead:  2,
			PermissionWrite: 0,
		},
	}

	if len(writeObjects) == 0 {
		logger.Debug("Can not generate deals for collection")
		return
	}

	_, err = nk.StorageWrite(ctx, writeObjects)
	if err != nil {
		logger.Error("Write exchange deals collection error %s", err.Error())
	}
}

func RpcExChangedealsList() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		exChangedealInShop, err := LoadExchangeDeals(ctx, logger, nk)
		if err != nil {
			logger.Error("Error when unmarshal list exchange deals, error %s", err.Error())
			return "", presenter.ErrUnmarshal
		}

		if exChangedealInShop == nil {
			return "", nil
		}

		marshaler := conf.MarshalerDefault
		exChangedealInShopJson, _ := marshaler.Marshal(exChangedealInShop)
		logger.Info("exchange deals results %s", exChangedealInShopJson)
		return string(exChangedealInShopJson), nil
	}
}

func LoadExchangeDeals(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) (*pb.ExchangeDealInShop, error) {
	objectIds := []*runtime.StorageRead{
		{
			Collection: kExchangeCollection,
			Key:        kExchangeKey,
		},
	}
	objects, err := nk.StorageRead(ctx, objectIds)
	exChangedealInShop := &pb.ExchangeDealInShop{}
	if err != nil {
		logger.Error("Error when read exchange deals , error %s", err.Error())
		return nil, presenter.ErrBetNotFound
	}
	if len(objects) == 0 {
		logger.Warn("List deals in storage empty")
		return exChangedealInShop, nil
	}

	unmarshaler := conf.Unmarshaler
	err = unmarshaler.Unmarshal([]byte(objects[0].GetValue()), exChangedealInShop)
	if err != nil {
		logger.Error("Error when unmarshal list bets, error %s", err.Error())
		return exChangedealInShop, presenter.ErrUnmarshal
	}
	return exChangedealInShop, nil
}

func RpcRequestNewExchange() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}
		unmarshaler := conf.Unmarshaler

		exChangedealReq := &pb.ExchangeInfo{}
		if err := unmarshaler.Unmarshal([]byte(payload), exChangedealReq); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		if exChangedealReq.CashId == "" || exChangedealReq.CashType == "" {
			logger.Error("Missing cash info")
			return "", errors.New("missing cash info")
		}
		// check valid id deal
		deal, err := GetExchangeDealFromId(exChangedealReq.Id)
		if err != nil {
			logger.Error("User %s request add new exchange with id %s, error: %s", userID, exChangedealReq.Id, err.Error())
			return "", presenter.ErrInternalError
		}
		profile, _, err := cgbdb.GetProfileUser(ctx, db, userID, nil)
		if err != nil {
			logger.Error("User %s get account info error %s", userID, err.Error())
			return "", presenter.ErrInternalError
		}
		exchange := &pb.ExchangeInfo{
			IdDeal:          deal.GetId(),
			Chips:           deal.Chips,
			Price:           deal.Price,
			Status:          int64(pb.ExchangeStatus_EXCHANGE_STATUS_WAITING.Number()),
			Unlock:          0,
			CashId:          exChangedealReq.CashId,
			CashType:        exChangedealReq.CashType,
			DeviceId:        exChangedealReq.DeviceId,
			UserIdRequest:   userID,
			UserNameRequest: profile.GetUserName(),
			VipLv:           profile.GetVipLevel(),
		}
		id, err := cgbdb.AddNewExchange(ctx, logger, db, exchange)
		if err != nil {
			logger.Error("AddNewExchange error %s", err.Error())
			return "", presenter.ErrInternalError
		}
		exchange.Id = id
		marshaler := conf.MarshalerDefault
		exChangeJson, _ := marshaler.Marshal(exchange)
		return string(exChangeJson), nil
	}
}

func GetExchangeDealFromId(id string) (*pb.Deal, error) {
	if deal, exist := MapExchangeDeal[id]; exist {
		return deal, nil
	}
	return nil, errors.New("Exchange id not found")
}

func RpcRequestCancelExchange() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}
		exChangedealReq := &pb.ExchangeInfo{}
		unmarshaler := conf.Unmarshaler
		if err := unmarshaler.Unmarshal([]byte(payload), exChangedealReq); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}
		if exChangedealReq.Id == "" {
			logger.Error("User %s query exchange %s is empty", userID, exChangedealReq.Id)
			return "", presenter.ErrInternalError
		}
		exChangedealReq.UserIdRequest = userID
		exChangeInDb, err := cgbdb.CancelExchangeByIdByUser(ctx, logger, db, exChangedealReq)
		if err != nil {
			logger.Error("User %s read exchange id %s error %s", userID, exChangedealReq.Id, err.Error())
			return "", presenter.ErrInternalError
		}
		marshaler := conf.MarshalerDefault
		exChangeInDbJson, _ := marshaler.Marshal(exChangeInDb)
		return string(exChangeInDbJson), nil
	}
}
