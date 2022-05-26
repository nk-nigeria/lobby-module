package api

import (
	"context"
	"database/sql"

	"github.com/ciaolink-game-platform/cgp-lobby-module/constant"
	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
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
		for _, validatePurchase := range listValidatePurchase {
			if validatePurchase.SeenBefore {
				logger.Warn("User %s , validate duplicate purchase %s", userID, validatePurchase.ProviderResponse)
				continue
			}
			if err := topupChipByIAP(ctx, logger, db, nk, userID, validatePurchase); err != nil {
				logger.Error("User %s, topup by IAP error %s , purchase %s", userID, err.Error(), validatePurchase.ProviderResponse)
				return err
			}
		}
		return nil
	}
	return x
}

func topupChipByIAP(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, userID string, purchasenk *nkapi.ValidatedPurchase) error {
	metadata := make(map[string]interface{})
	metadata["action"] = "iap_topup"
	metadata["sender"] = constant.UUID_USER_SYSTEM
	metadata["recv"] = userID
	metadata["iap_type"] = "google"
	wallet := entity.Wallet{
		UserId: userID,
		Chips:  100,
	}
	err := entity.AddChipWalletUser(ctx, nk, logger, userID, wallet, metadata)
	return err
}
