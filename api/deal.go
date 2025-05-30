package api

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	pb "github.com/nakamaFramework/cgp-common/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	kDealCollection = "deal-collection"
	kDealKey        = "deal-key"
)

var MapDeal = make(map[string]*pb.Deal)

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
		dealInShop := &pb.DealInShop{}
		_ = conf.Unmarshaler.Unmarshal([]byte(objects[0].GetValue()), dealInShop)
		MapDeal[dealInShop.Best.Id] = dealInShop.Best
		for _, deal := range dealInShop.Gcashes {
			MapDeal[deal.GetId()] = deal
		}
		for _, deal := range dealInShop.Iaps {
			MapDeal[deal.GetId()] = deal
		}
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
			Currency:    "$",
			Percent:     "15.5",
		},
		Iaps: []*pb.Deal{
			{
				Id:          "id_best_iap_1",
				Chips:       100000,
				Bonus:       0,
				Price:       "10",
				AmountChips: 100000,
				Name:        "100,000 Chips",
				Currency:    "$",
				Percent:     "25",
			},
			{
				Id:          "id_best_iap_2",
				Chips:       300000,
				Bonus:       0,
				Price:       "20",
				AmountChips: 300000,
				Name:        "300,000 Chips",
				Currency:    "$",
				Percent:     "88",
			},
			{
				Id:          "id_best_iap_3",
				Chips:       750000,
				Bonus:       0,
				Price:       "50",
				AmountChips: 750000,
				Name:        "750,000 Chips",
				Currency:    "$",
				Percent:     "88",
			},
			{
				Id:          "id_best_iap_4",
				Chips:       1700000,
				Bonus:       0,
				Price:       "100",
				AmountChips: 1700000,
				Name:        "1,700,000 Chips",
				Currency:    "$",
				Percent:     "113",
			},
			{
				Id:          "id_best_iap_5",
				Chips:       3600000,
				Bonus:       0,
				Price:       "200",
				AmountChips: 3600000,
				Name:        "3,600,000 Chips",
				Currency:    "$",
				Percent:     "125",
			},
			// {
			// 	Id:          "id_best_iap_6",
			// 	Chips:       9000000,
			// 	Bonus:       0,
			// 	Price:       "500",
			// 	AmountChips: 9000000,
			// 	Name:        "9,000,000 Chips",
			// 	Currency:    "$",
			// 	Percent:     "125",
			// },
			// {
			// 	Id:          "id_best_iap_7",
			// 	Chips:       18500000,
			// 	Bonus:       0,
			// 	Price:       "1000",
			// 	AmountChips: 18500000,
			// 	Name:        "18,500,000 Chips",
			// 	Currency:    "$",
			// 	Percent:     "131",
			// },
			// {
			// 	Id:          "id_best_iap_8",
			// 	Chips:       38000000,
			// 	Bonus:       0,
			// 	Price:       "2000",
			// 	AmountChips: 38000000,
			// 	Name:        "38,000,000 Chips",
			// 	Currency:    "$",
			// 	Percent:     "138",
			// },
		},
		Gcashes: []*pb.Deal{
			{
				Id:          "id_best_gcash1",
				Chips:       3000,
				Bonus:       60,
				Price:       "1000",
				AmountChips: 1050,
				Name:        "Best deal",
				Currency:    "$",
				Percent:     "15.5",
			},
			{
				Id:          "id_best_gcash2",
				Chips:       5000,
				Bonus:       70,
				Price:       "2000",
				AmountChips: 1050,
				Name:        "Best deal",
				Currency:    "$",
				Percent:     "15.5",
			},
		},
		Sms: []*pb.Deal{
			{
				Id:          "id_best_sms_1",
				Chips:       3,
				Bonus:       90,
				Price:       "20",
				AmountChips: 1050,
				Name:        "Best deal",
				Currency:    "$",
				Percent:     "15.5",
			},
			{
				Id:          "id_best_sms_2",
				Chips:       5,
				Bonus:       1200,
				Price:       "100",
				AmountChips: 1050,
				Name:        "Best deal",
				Currency:    "$",
				Percent:     "15.5",
			},
		},
	}

	for _, deal := range deals.GetIaps() {
		priceInt, _ := strconv.ParseInt(deal.GetPrice(), 10, 64)
		deal.ChipPerUnit = deal.GetChips() / priceInt
	}
	for _, deal := range deals.GetGcashes() {
		priceInt, _ := strconv.ParseInt(deal.GetPrice(), 10, 64)
		deal.ChipPerUnit = deal.GetChips() / priceInt
	}
	{
		priceInt, _ := strconv.ParseInt(deals.GetBest().GetPrice(), 10, 64)
		deals.GetBest().ChipPerUnit = deals.GetBest().GetChips() / priceInt
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
	MapDeal[deals.Best.Id] = deals.Best
	for _, deal := range deals.Gcashes {
		MapDeal[deal.GetId()] = deal
	}
	for _, deal := range deals.Iaps {
		MapDeal[deal.GetId()] = deal
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
			logger.Error("Error when unmarshal list deals, error %s", err.Error())
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
		logger.Error("Error when unmarshal list deals, error %s", err.Error())
		return dealInShop, presenter.ErrUnmarshal
	}
	return dealInShop, nil
}
