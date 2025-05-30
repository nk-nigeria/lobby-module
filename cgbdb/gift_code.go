package cgbdb

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgtype"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	"github.com/nakamaFramework/cgb-lobby-module/constant"
	pb "github.com/nakamaFramework/cgp-common/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CREATE TABLE
//   public.giftcode (
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
//   public.giftcode
// ADD
//   CONSTRAINT giftcode_pkey PRIMARY KEY (id)

const GiftCodeTableName = "giftcode"

func AddNewGiftCode(ctx context.Context, logger runtime.Logger, db *sql.DB, giftCode *pb.GiftCode) (*pb.GiftCode, error) {
	if giftCode == nil || giftCode.GetCode() == "" || giftCode.GetValue() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Error add giftcode.")
	}
	if giftCode.GetValue() > constant.MaxChipAllowAdd {
		return nil, status.Error(codes.OutOfRange, "giftcode value too large")
	}
	giftCode.Id = conf.SnowlakeNode.Generate().Int64()
	query := "INSERT INTO " + GiftCodeTableName + " (id, code, n_current, n_max, value, start_time_unix, end_time_unix, message, vip, gift_code_type, create_time, update_time) VALUES ($1, $2, $3, $4, $5, to_timestamp($6), to_timestamp($7), $8, $9, $10, now(), now())"
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
			return nil, status.Error(codes.Internal, "Duplicate giftcode")
		}
		return nil, status.Error(codes.Internal, "Error add giftcode")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert new gift code %s",
			giftCode.GetCode())
		return nil, status.Error(codes.Internal, "Error add giftcode.")
	}
	return GetGiftCode(ctx, logger, db, giftCode)
}

func GetGiftCode(ctx context.Context, logger runtime.Logger, db *sql.DB, giftCode *pb.GiftCode) (*pb.GiftCode, error) {
	if giftCode == nil || giftCode.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "Error query giftcode.")
	}
	query := "SELECT id, code, n_current, n_max, value, start_time_unix, end_time_unix, message, vip, gift_code_type, create_time, update_time FROM " + GiftCodeTableName + " WHERE code=$1"
	var dbID int64
	var dbCode, dbMessage string
	var dbNCurrent, dbNMax, dbValue, dbVip, dbGiftCodeType int64
	var dbStartTime, dbEndTime, dbCreateTime, dbUpdateTime pgtype.Timestamptz
	err := db.QueryRowContext(ctx, query, giftCode.Code).Scan(
		&dbID, &dbCode, &dbNCurrent,
		&dbNMax, &dbValue, &dbStartTime,
		&dbEndTime, &dbMessage, &dbVip,
		&dbGiftCodeType, &dbCreateTime, &dbUpdateTime)
	if err != nil {
		logger.Error("Query giftcode %s,  error %s",
			giftCode.GetCode(), err.Error())
		return nil, status.Error(codes.Internal, "Query giftcode error")
	}
	respGiftCode := pb.GiftCode{
		Id:            dbID,
		Code:          dbCode,
		NCurrent:      dbNCurrent,
		NMax:          dbNMax,
		Value:         dbValue,
		StartTimeUnix: dbStartTime.Time.Unix(),
		EndTimeUnix:   dbEndTime.Time.Unix(),
		Message:       dbMessage,
		Vip:           dbVip,
		GiftCodeType:  pb.GiftCodeType(dbGiftCodeType),
	}
	if respGiftCode.GetNCurrent() == respGiftCode.GetNMax() {
		respGiftCode.ReachMaxClaim = true
	}
	return &respGiftCode, nil
}

func GetListGiftCode(ctx context.Context, logger runtime.Logger, db *sql.DB, giftCode *pb.GiftCode) (pb.ListGiftCode, error) {
	query := "SELECT id, code, n_current, n_max, value, start_time_unix, end_time_unix, message, vip, gift_code_type, create_time, update_time FROM " + GiftCodeTableName + " order by id desc"
	var dbID int64
	var dbCode, dbMessage string
	var dbNCurrent, dbNMax, dbValue, dbVip, dbGiftCodeType int64
	var dbStartTime, dbEndTime, dbCreateTime, dbUpdateTime pgtype.Timestamptz
	params := make([]interface{}, 0)
	// params = append(params, giftCode.GetCode())
	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		logger.Error("Query giftcode %s,  error %s",
			giftCode.GetCode(), err.Error())
		return pb.ListGiftCode{
			GiftCodes: make([]*pb.GiftCode, 0, 0),
		}, status.Error(codes.Internal, "Query giftcode error")
	}
	ml := make([]*pb.GiftCode, 0)
	for rows.Next() {
		rows.Scan(
			&dbID, &dbCode, &dbNCurrent,
			&dbNMax, &dbValue, &dbStartTime,
			&dbEndTime, &dbMessage, &dbVip,
			&dbGiftCodeType, &dbCreateTime, &dbUpdateTime)
		respGiftCode := pb.GiftCode{
			Id:            dbID,
			Code:          dbCode,
			NCurrent:      dbNCurrent,
			NMax:          dbNMax,
			Value:         dbValue,
			StartTimeUnix: dbStartTime.Time.Unix(),
			EndTimeUnix:   dbEndTime.Time.Unix(),
			Message:       dbMessage,
			Vip:           dbVip,
			GiftCodeType:  pb.GiftCodeType(dbGiftCodeType),
		}
		if respGiftCode.GetNCurrent() == respGiftCode.GetNMax() {
			respGiftCode.ReachMaxClaim = true
		}
		ml = append(ml, &respGiftCode)
	}
	sort.Slice(ml, func(i, j int) bool {
		return ml[i].Id > ml[j].Id
	})
	return pb.ListGiftCode{
		GiftCodes: ml,
	}, nil
}

func ClaimGiftCode(ctx context.Context, logger runtime.Logger, db *sql.DB, giftCode *pb.GiftCode, lvVipUser int64) (*pb.GiftCode, error) {
	if giftCode == nil || giftCode.GetCode() == "" || giftCode.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "Error query giftcode.")
	}
	dbGiftCode, err := GetGiftCode(ctx, logger, db, giftCode)
	if err != nil {
		return nil, err
	}

	dbGiftCode.UserId = giftCode.GetUserId()
	nowUnix := time.Now().Unix()
	dbGiftCode.OpenToClaim = true
	if dbGiftCode.StartTimeUnix > nowUnix {
		dbGiftCode.OpenToClaim = false
		dbGiftCode.ErrCode = int32(pb.GiftCodeError_GIFT_CODE_ERROR_NOT_OPEN)
		logger.Error("Giftcode %s not open", dbGiftCode.Code)
		return dbGiftCode, nil
	}
	if dbGiftCode.EndTimeUnix < nowUnix {
		dbGiftCode.OpenToClaim = false
		dbGiftCode.ErrCode = int32(pb.GiftCodeError_GIFT_CODE_ERROR_HAS_CLOSED)
		logger.Error("Giftcode %s close", dbGiftCode.Code)
		return dbGiftCode, nil
	}

	if dbGiftCode.NCurrent >= dbGiftCode.GetNMax() {
		dbGiftCode.ReachMaxClaim = true
		dbGiftCode.ErrCode = int32(pb.GiftCodeError_GIFT_CODE_ERROR_REACH_MAX_CLAIMED)
		return dbGiftCode, nil
	}

	if dbGiftCode.Vip > lvVipUser {
		dbGiftCode.ErrCode = int32(pb.GiftCodeError_GIFT_CODE_ERROR_LV_VIP_NOT_MEET_REQUIRE)
		return dbGiftCode, nil
	}

	giftCodeClaim, err := GetGiftCodeClaim(ctx, logger, db, dbGiftCode)
	if err != nil {
		logger.Error("GetGiftCodeClaim user %s error %s", dbGiftCode.GetUserId(), err.Error())
		return nil, status.Error(codes.InvalidArgument, "Error query giftcodeclaim.")

	} else if giftCodeClaim.Code != "" {
		dbGiftCode.AlreadyClaim = true
		dbGiftCode.ErrCode = int32(pb.GiftCodeError_GIFT_CODE_ERROR_ALREADY_CLAIMED)
		return dbGiftCode, nil
	}

	queryCheckCodeClaimByUser := "Select code from " + GiftCodeClaimTableName + " where user_id=$2 AND id_code=$3 AND code=$4"
	query := `UPDATE ` + GiftCodeTableName + " SET n_current=n_current+1, update_time=now() where code=$1 AND n_current<n_max AND code NOT IN ( " + queryCheckCodeClaimByUser + " )"
	result, err := db.ExecContext(ctx, query, dbGiftCode.Code, dbGiftCode.GetUserId(), dbGiftCode.GetId(), dbGiftCode.GetCode())
	if err != nil {
		logger.Error("Cannot claim giftcode %s, err: %s",
			giftCode.GetCode(), err.Error())
		return nil, status.Error(codes.Internal, "Error claim giftcode")
	}

	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not update gift code claim.")
		return nil, status.Error(codes.Internal, " Error claim giftcode.")
	}
	dbGiftCode.NCurrent++
	AddNewGiftCodeClaim(ctx, logger, db, dbGiftCode)
	return dbGiftCode, nil
}

func DeletedGiftCode(ctx context.Context, logger runtime.Logger, db *sql.DB, giftCode *pb.GiftCode) (*pb.GiftCode, error) {
	dbGiftCode, err := GetGiftCode(ctx, logger, db, giftCode)
	if err != nil {
		logger.Error("GetGiftCode code %d before delete error %s", giftCode.GetCode(), err.Error())
		return nil, err
	}
	_, err = AddGiftCodeTombstone(ctx, logger, db, dbGiftCode)
	if err != nil {
		logger.Error("AddGiftCodeTombstone code %s error %s", giftCode.GetCode(), err.Error())
		return nil, err
	}

	query := `DELETE FROM ` + GiftCodeTableName + " WHERE code=$1"
	result, err := db.ExecContext(ctx, query, giftCode.Code)
	if err != nil {
		logger.Error("Cannot deleted giftcode %s, err: %s",
			giftCode.GetCode(), err.Error())
		return nil, status.Error(codes.Internal, "Error deleted giftcode")
	}

	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not update new user.")
		return nil, status.Error(codes.Internal, "Error deleted giftcode.")
	}
	return dbGiftCode, nil
}
