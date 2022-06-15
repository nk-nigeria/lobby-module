package cgbdb

import (
	"context"
	"database/sql"
	"strings"

	"github.com/ciaolink-game-platform/cgp-lobby-module/conf"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CREATE TABLE
//   public.giftcodeclaim (
//      id bigint NOT NULL,
//     code character varying(128) NOT NULL DEFAULT '',
//     user_id character varying(128) NOT NULL,
//     create_time timestamp
//     with
//       time zone NOT NULL DEFAULT now(),
//       update_time timestamp
//     with
//       time zone NOT NULL DEFAULT now()
//   );

// ALTER TABLE
//   public.giftcodeclaim
// ADD
//   CONSTRAINT giftcodeclaim_pkey PRIMARY KEY (id)
const GiftCodeClaimTableName = "giftcodeclaim"

func AddNewGiftCodeClaim(ctx context.Context, logger runtime.Logger, db *sql.DB, giftCode *pb.GiftCode) error {
	if giftCode == nil || giftCode.GetCode() == "" || giftCode.GetValue() <= 0 {

	}
	query := "INSERT INTO " + GiftCodeClaimTableName + " (id, code, user_id, create_time, update_time) VALUES ($1, $2, $3, now(), now())"
	// startTime := pgtype.Timestamptz
	result, err := db.ExecContext(ctx, query,
		conf.SnowlakeNode.Generate().Int64(), giftCode.GetCode(), giftCode.GetUserId())
	if err != nil {
		logger.Error("Add new giftcodeclaim %s, error %s",
			giftCode.GetCode(), err.Error())

		return status.Error(codes.Internal, "Error add giftcodeclaim")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert new gift code %s",
			giftCode.GetCode())
		return status.Error(codes.Internal, "Error add giftcodeclaim.")
	}
	return nil
}

func GetGiftCodeClaim(ctx context.Context, logger runtime.Logger, db *sql.DB, giftCode *pb.GiftCode) (*pb.GiftCode, error) {
	if giftCode == nil || giftCode.GetCode() == "" || giftCode.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Error query giftcode.")
	}
	query := "SELECT id, code, user_id, create_time, update_time FROM " + GiftCodeClaimTableName + " WHERE user_id=$1 AND code=$2"
	var dbID int64
	var dbCode, dbUserId string
	var dbCreateTime, dbUpdateTime pgtype.Timestamptz
	err := db.QueryRowContext(ctx, query, giftCode.GetUserId(), giftCode.GetCode()).Scan(
		&dbID, &dbCode, &dbUserId,
		&dbCreateTime, &dbUpdateTime)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return &pb.GiftCode{}, nil
		}
		logger.Error("Query giftcode %s,  error %s",
			giftCode.GetCode(), err.Error())
		return nil, status.Error(codes.Internal, "Query giftcode error")
	}
	respGiftCode := pb.GiftCode{
		Id:     dbID,
		Code:   dbCode,
		UserId: dbUserId,
	}

	return &respGiftCode, nil
}
