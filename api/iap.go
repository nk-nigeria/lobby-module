package api

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/ciaolink-game-platform/cgb-lobby-module/constant"
	"github.com/ciaolink-game-platform/cgb-lobby-module/entity"
	nkapi "github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RegisterValidatePurchase(db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) {
	initializer.RegisterAfterValidatePurchaseGoogle(validatePurchaseGoogle())
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
			if err := topupChipByIAP(ctx, logger, db, nk, userID, validatePurchase); err != nil {
				logger.Error("User %s, topup by IAP error %s , purchase %s", userID, err.Error(), validatePurchase.ProviderResponse)
				return err
			}
			logger.Info("[Success] Top up user %s ,from product id %s, transID %s ", userID, validatePurchase.ProductId, validatePurchase.TransactionId)
		}
		return nil
	}
	return x
}

func topupChipByIAP(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, userID string, purchasenk *nkapi.ValidatedPurchase) error {
	metadata := make(map[string]interface{})
	metadata["action"] = entity.WalletActionIAPTopUp
	metadata["sender"] = constant.UUID_USER_SYSTEM
	metadata["recv"] = userID
	metadata["iap_type"] = "google"
	metadata["trans_id"] = purchasenk.TransactionId
	deal, exits := MapDeal[purchasenk.ProductId]
	if !exits {
		logger.Error("Get deal from product id %s error: Not found", purchasenk.ProductId)
		return errors.New("product id not found")
	}
	wallet := entity.Wallet{
		UserId: userID,
		Chips:  deal.Chips,
	}
	err := entity.AddChipWalletUser(ctx, nk, logger, userID, wallet, metadata)
	return err
}
