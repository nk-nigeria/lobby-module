package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ciaolink-game-platform/cgb-lobby-module/cgbdb"
	"github.com/ciaolink-game-platform/cgb-lobby-module/constant"
	"github.com/ciaolink-game-platform/cgb-lobby-module/entity"
	nkapi "github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IAPType string

const (
	IAP_SYSTEM IAPType = "system"
	IAP_GOOGLE IAPType = "google"
)

type IAPRequest struct {
	UserId    string `json:"user_id,omitempty"`
	ProductId string `json:"product_id,omitempty"`
}

func RegisterValidatePurchase(db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) {
	initializer.RegisterAfterValidatePurchaseGoogle(validatePurchaseGoogle())
}

func RpcIAP() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if userID != "" {
			return "", errors.New("Unauth")
		}
		iapReq := IAPRequest{}
		err := json.Unmarshal([]byte(payload), &iapReq)
		if err != nil {
			return "", err
		}
		if iapReq.UserId == "" {
			return "", errors.New("missing user id")
		}
		transaction := fmt.Sprintf("trans-%s", time.Now().String())
		err = topupChipIAP(ctx, logger, db, nk, iapReq.UserId, IAP_SYSTEM, transaction, iapReq.ProductId)
		return "", err
	}
}
func validatePurchaseGoogle() func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out *nkapi.ValidatePurchaseResponse, in *nkapi.ValidatePurchaseGoogleRequest) error {
	x := func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out *nkapi.ValidatePurchaseResponse, in *nkapi.ValidatePurchaseGoogleRequest) error {

		if out == nil {
			logger.Error("Invalid validate purchase, out is nil")
			return status.Error(codes.InvalidArgument, "out is nil")
		}
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			logger.Error("context did not contain user ID.")
			return status.Error(codes.InvalidArgument, "user id not found")
		}
		listValidatePurchase := out.GetValidatedPurchases()
		productIDs := make([]string, len(listValidatePurchase))
		for _, p := range listValidatePurchase {
			productIDs = append(productIDs, p.GetProductId())
		}
		logger.Info("validatePurchaseGoogle userId %s, purchase id %s", userID, strings.Join(productIDs, ","))

		for _, validatePurchase := range listValidatePurchase {
			if validatePurchase.SeenBefore {
				logger.Warn("User %s , validate duplicate purchase %s", userID, validatePurchase.ProviderResponse)
				continue
			}
			if err := topupChipIAPByPurchase(ctx, logger, db, nk, userID, validatePurchase); err != nil {
				logger.Error("User %s, topup by IAP error %s , purchase %s", userID, err.Error(), validatePurchase.ProviderResponse)
				return err
			}
			logger.Info("[Success] Top up user %s ,from product id %s, transID %s ", userID, validatePurchase.ProductId, validatePurchase.TransactionId)
		}
		return nil
	}
	return x
}

func topupChipIAPByPurchase(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, userID string, purchasenk *nkapi.ValidatedPurchase) error {
	err := topupChipIAP(ctx, logger, db, nk, userID, IAP_GOOGLE, purchasenk.TransactionId, purchasenk.ProductId)
	return err
}

func topupChipIAP(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, userID string, typeIAP IAPType, transactionId string, productId string) error {
	metadata := make(map[string]interface{})
	metadata["action"] = entity.WalletActionIAPTopUp
	metadata["sender"] = constant.UUID_USER_SYSTEM
	metadata["recv"] = userID
	metadata["iap_type"] = string(typeIAP)
	metadata["trans_id"] = transactionId
	deal, exits := MapDeal[productId]
	if !exits {
		logger.WithField("product id", productId).Error("Get deal from product failed")
		return errors.New("product id not found")
	}
	wallet := entity.Wallet{
		UserId: userID,
		Chips:  deal.Chips,
	}
	err := entity.AddChipWalletUser(ctx, nk, logger, userID, wallet, metadata)
	if err == nil {
		cgbdb.UpdateTopupSummary(db, userID, deal.Chips)
	}
	return err
}
