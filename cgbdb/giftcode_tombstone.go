package cgbdb

import (
	"context"
	"database/sql"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
	pb "github.com/nakamaFramework/cgp-common/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CREATE TABLE
//   public.giftcodetombstone (
//      id bigint NOT NULL,
//     code character varying(128) NOT NULL DEFAULT '',
//     UNIQUE(code),
//     n_current integer NOT NULL DEFAULT 0,
//     n_max integer NOT NULL DEFAULT 0,
//     value integer NOT NULL DEFAULT 0,
//     start_time_unix timestamp,
//     end_time_unix timestamp,
//     message character varying(256) NOT NULL DEFAULT '',
//     vip integer NOT NULL DEFAULT 0,
//     gift_code_type smallint NOT NULL DEFAULT 1,
//     create_time timestamp

//     with
//       time zone NOT NULL DEFAULT now(),
//       update_time timestamp
//     with
//       time zone NOT NULL DEFAULT now()
//   );

// ALTER TABLE
//
//	public.giftcodetombstone
//
// ADD
//
//	CONSTRAINT giftcodetombstone_pkey PRIMARY KEY (id)
const GiftCodeTombstoneTableName = "giftcodetombstone"

func AddGiftCodeTombstone(ctx context.Context, logger runtime.Logger, db *sql.DB, giftCode *pb.GiftCode) (*pb.GiftCode, error) {
	if giftCode == nil || giftCode.GetCode() == "" || giftCode.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Error add giftcodetombstone.")
	}
	query := "INSERT INTO " + GiftCodeTombstoneTableName + " (id, code, n_current, n_max, value, start_time_unix, end_time_unix, message, vip, gift_code_type, create_time, update_time) VALUES ($1, $2, $3, $4, $5, to_timestamp($6), to_timestamp($7), $8, $9, $10, now(), now())"
	// startTime := pgtype.Timestamptz
	result, err := db.ExecContext(ctx, query,
		giftCode.GetId(), giftCode.GetCode(), giftCode.GetNCurrent(),
		giftCode.GetNMax(), giftCode.GetValue(), giftCode.GetStartTimeUnix(),
		giftCode.GetEndTimeUnix(), giftCode.GetMessage(), giftCode.GetVip(),
		giftCode.GetGiftCodeType().Number())
	if err != nil {
		logger.Error("Add new giftcode %s, error %s",
			giftCode.GetCode(), err.Error())
		if strings.Contains(err.Error(), "duplicate") {
			return nil, status.Error(codes.Internal, "Duplicate giftcodetombstone")
		}
		return nil, status.Error(codes.Internal, "Error add giftcodetombstone")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert new gift code tombstone %s",
			giftCode.GetCode())
		return nil, status.Error(codes.Internal, "Error add giftcode.")
	}
	return GetGiftCode(ctx, logger, db, giftCode)
}
