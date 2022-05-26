package api

import (
	"context"
	"database/sql"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	kDealCollection = "deal-collection"
	kDealKey        = "deal-key"
)

func InitDeal(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, marshaler *protojson.MarshalOptions) {
	objectIds := []*runtime.StorageRead{
		{
			Collection: kDealCollection,
			Key:        kDealKey,
		},
	}

	objects, err := nk.StorageRead(ctx, objectIds)
	if err != nil {
		logger.Error("Error when read deal at init, error %s", err.Error())
	}

	if len(objects) > 0 {
		logger.Info("List deal already write in collection")
		return
	}
	deals := pb.DealInShop{
		Best: &pb.Deal{
			Id:          "id_best_deal",
			Chips:       1000,
			Bonus:       50,
			Price:       "2000",
			AmountChips: 1050,
			Name:        "Best deal",
			Currency:    "VND",
			Percent:     "15.5",
		},
		Iaps: []*pb.Deal{
			{
				Id:          "id_best_iap_1",
				Chips:       100,
				Bonus:       50,
				Price:       "1000 Vnd",
				AmountChips: 1050,
				Name:        "Best deal",
				Currency:    "VND",
				Percent:     "15.5",
			},
			{
				Id:          "id_best_iap_2",
				Chips:       200,
				Bonus:       60,
				Price:       "2000 Vnd",
				AmountChips: 1050,
				Name:        "Best deal",
				Currency:    "VND",
				Percent:     "15.5",
			},
		},
		Gcashes: []*pb.Deal{
			{
				Id:          "id_best_gcash1",
				Chips:       3000,
				Bonus:       60,
				Price:       "1000 Vnd",
				AmountChips: 1050,
				Name:        "Best deal",
				Currency:    "VND",
				Percent:     "15.5",
			},
			{
				Id:          "id_best_gcash2",
				Chips:       5000,
				Bonus:       70,
				Price:       "2000 Vnd",
				AmountChips: 1050,
				Name:        "Best deal",
				Currency:    "VND",
				Percent:     "15.5",
			},
		},
		Sms: []*pb.Deal{
			{
				Id:          "id_best_sms_1",
				Chips:       3,
				Bonus:       90,
				Price:       "20 Vnd",
				AmountChips: 1050,
				Name:        "Best deal",
				Currency:    "VND",
				Percent:     "15.5",
			},
			{
				Id:          "id_best_sms_2",
				Chips:       5,
				Bonus:       100,
				Price:       "100 Vnd",
				AmountChips: 1050,
				Name:        "Best deal",
				Currency:    "VND",
				Percent:     "15.5",
			},
		},
	}
	dealsJson, err := marshaler.Marshal(&deals)
	if err != nil {
		logger.Debug("Can not marshaler deals for collection")
		return
	}

	writeObjects := []*runtime.StorageWrite{
		{
			Collection:      kDealCollection,
			Key:             kDealKey,
			Value:           string(dealsJson),
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
		logger.Error("Write deals collection error %s", err.Error())
	}
}

func RpcDealList(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		dealInShop, err := LoadDeals(ctx, logger, nk, unmarshaler)
		if err != nil {
			logger.Error("Error when unmarshal list bets, error %s", err.Error())
			return "", presenter.ErrUnmarshal
		}

		if dealInShop == nil {
			return "", nil
		}

		dealInShopJson, _ := marshaler.Marshal(dealInShop)
		logger.Info("bets results %s", dealInShopJson)
		return string(dealInShopJson), nil
	}
}

func LoadDeals(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, unmarshaler *protojson.UnmarshalOptions) (*pb.DealInShop, error) {
	objectIds := []*runtime.StorageRead{
		{
			Collection: kDealCollection,
			Key:        kDealKey,
		},
	}
	objects, err := nk.StorageRead(ctx, objectIds)
	dealInShop := &pb.DealInShop{}
	if err != nil {
		logger.Error("Error when read deals , error %s", err.Error())
		return nil, presenter.ErrBetNotFound
	}
	if len(objects) == 0 {
		logger.Warn("List deals in storage empty")
		return dealInShop, nil
	}

	err = unmarshaler.Unmarshal([]byte(objects[0].GetValue()), dealInShop)
	if err != nil {
		logger.Error("Error when unmarshal list bets, error %s", err.Error())
		return dealInShop, presenter.ErrUnmarshal
	}
	return dealInShop, nil
}
