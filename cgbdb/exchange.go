package cgbdb

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgtype"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	"github.com/nakamaFramework/cgb-lobby-module/entity"
	pb "github.com/nakamaFramework/cgp-common/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const ExchangeTableName = "exchange"

// CREATE TABLE public.exchange (
//   id bigint NOT NULL,
//   id_deal character varying(128) NOT NULL,
// 	 chips integer NOT NULL DEFAULT 0,
//   price character varying(128) NOT NULL,
//   status smallint NOT NULL DEFAULT 0,
//   unlock smallint NOT NULL DEFAULT 1,
//   cash_id character varying(128) NOT NULL,
//   cash_type character varying(128) NOT NULL,
//   user_id_request character varying(128) NOT NULL,
//   user_name_request character varying(128) NOT NULL,
//   vip_lv smallint NOT NULL DEFAULT 0,
//   device_id character varying(128) NOT NULL,
//   user_id_handling character varying(128) NOT NULL,
//   user_name_handling character varying(128) NOT NULL,
//   reason character varying(128) NOT NULL,
//   create_time timestamp with time zone NOT NULL DEFAULT now(),
//   update_time timestamp with time zone NOT NULL DEFAULT now()
// );
// ALTER TABLE
//   public.exchange
// ADD
//   CONSTRAINT exchange_pkey PRIMARY KEY (id)

func AddNewExchange(ctx context.Context, logger runtime.Logger, db *sql.DB, exchange *pb.ExchangeInfo) (string, error) {
	exchange.Id = conf.SnowlakeNode.Generate().String()
	query := "INSERT INTO " + ExchangeTableName +
		" (id, id_deal, chips, price, status, unlock, cash_id, cash_type, user_id_request, user_name_request, vip_lv, device_id, user_id_handling, user_name_handling, reason, create_time, update_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, now(), now())"
	result, err := db.ExecContext(ctx, query,
		exchange.Id, exchange.GetIdDeal(), exchange.GetChips(),
		exchange.GetPrice(), exchange.GetStatus(), 1,
		exchange.GetCashId(), exchange.GetCashType(), exchange.GetUserIdRequest(),
		exchange.GetUserNameRequest(), exchange.GetVipLv(), exchange.GetDeviceId(),
		exchange.GetUserIdHandling(), exchange.GetUserNameHandling(), exchange.GetReason())
	if err != nil {
		logger.Error("Error when add new exchange, user request: %s, chips: %d, price %s,  error %s",
			exchange.UserIdRequest, exchange.Chips, exchange.Price, err.Error())
		return "", status.Error(codes.Internal, "Error add exchange.")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert new exchange, user request: %s, chips: %d, price %s",
			exchange.UserIdRequest, exchange.Chips, exchange.Price)
		return "", status.Error(codes.Internal, "Error add exchange.")
	}
	return exchange.Id, nil
}

func GetExchangeByIdByUserID(ctx context.Context, logger runtime.Logger, db *sql.DB, exchange *pb.ExchangeInfo) (*pb.ExchangeInfo, error) {
	query := "SELECT id, chips, price, status, unlock, cash_id, cash_type, user_id_request, user_name_request, vip_lv, device_id, user_id_handling, user_name_handling, reason FROM " +
		ExchangeTableName + " WHERE id=$1 AND user_id_request=$2"
	var dbChips, dbStatus, dbVipLv int64
	var dbId, dbPrice, dbCashId, dbCashType, dbUserIdReq, dbUserNameReq,
		dbDeviceId, dbUserIdHandling, dbUserNameHandling, dbReason string
	var dbUnlock int32
	err := db.QueryRowContext(ctx, query, exchange.Id, exchange.GetUserIdRequest()).Scan(&dbId, &dbChips, &dbPrice, &dbStatus,
		&dbUnlock, &dbCashId, &dbCashType, &dbUserIdReq, &dbUserNameReq, &dbVipLv, &dbDeviceId, &dbUserIdHandling, &dbUserNameHandling, &dbReason)
	if err != nil {
		logger.Error("Query exchange id %s, user id %s, error %s", exchange.Id, exchange.UserIdRequest, err.Error())
		return nil, status.Error(codes.Internal, "Query exchange error")
	}
	resp := pb.ExchangeInfo{
		Id:               dbId,
		Chips:            dbChips,
		Price:            dbPrice,
		Status:           dbStatus,
		Unlock:           dbUnlock,
		CashId:           dbCashId,
		CashType:         dbCashType,
		UserIdRequest:    dbUserIdReq,
		UserNameRequest:  dbUserNameReq,
		VipLv:            dbVipLv,
		DeviceId:         dbDeviceId,
		UserIdHandling:   dbUserIdHandling,
		UserNameHandling: dbUserNameHandling,
		Reason:           dbReason,
	}
	return &resp, nil
}

func GetExchangeById(ctx context.Context, logger runtime.Logger, db *sql.DB, exchange *pb.ExchangeInfo) (*pb.ExchangeInfo, error) {
	query := "SELECT id, chips, price, status, unlock, cash_id, cash_type, user_id_request, user_name_request, vip_lv, device_id, user_id_handling, user_name_handling, reason FROM " +
		ExchangeTableName + " WHERE id=$1"
	var dbChips, dbStatus, dbVipLv int64
	var dbId, dbPrice, dbCashId, dbCashType, dbUserIdReq, dbUserNameReq,
		dbDeviceId, dbUserIdHandling, dbUserNameHandling, dbReason string
	var dbUnlock int32
	err := db.QueryRowContext(ctx, query, exchange.Id).Scan(
		&dbId, &dbChips, &dbPrice,
		&dbStatus, &dbUnlock, &dbCashId,
		&dbCashType, &dbUserIdReq, &dbUserNameReq,
		&dbVipLv, &dbDeviceId, &dbUserIdHandling,
		&dbUserNameHandling, &dbReason)
	if err != nil {
		logger.Error("Query exchange id %s, user id %s, error %s", exchange.Id, exchange.UserIdRequest, err.Error())
		return nil, status.Error(codes.Internal, "Query exchange error")
	}
	resp := pb.ExchangeInfo{
		Id:               dbId,
		Chips:            dbChips,
		Price:            dbPrice,
		Status:           dbStatus,
		Unlock:           dbUnlock,
		CashId:           dbCashId,
		CashType:         dbCashType,
		UserIdRequest:    dbUserIdReq,
		UserNameRequest:  dbUserNameReq,
		VipLv:            dbVipLv,
		DeviceId:         dbDeviceId,
		UserIdHandling:   dbUserIdHandling,
		UserNameHandling: dbUserNameHandling,
		Reason:           dbReason,
	}
	return &resp, nil
}

func CancelExchangeByIdByUser(ctx context.Context, logger runtime.Logger, db *sql.DB, exchange *pb.ExchangeInfo) (*pb.ExchangeInfo, error) {
	exChangeInDb, err := GetExchangeByIdByUserID(ctx, logger, db, exchange)
	if err != nil {
		return nil, err
	}
	if exChangeInDb.Unlock == 0 {
		logger.Error("User %s request cancel exchange id %s error: exchange had locked",
			exChangeInDb.GetUserIdRequest(), exChangeInDb.GetId())
		return exChangeInDb, nil
	}
	if exChangeInDb.Status != int64(pb.ExchangeStatus_EXCHANGE_STATUS_WAITING.Number()) {
		logger.Error("User %s request cancel exchange id %s error: status not waiting",
			exChangeInDb.GetUserIdRequest(), exChangeInDb.GetId())
		return exChangeInDb, nil
	}
	query := "UPDATE " + ExchangeTableName + " SET status=$1 WHERE id=$2 AND user_id_request=$3 AND status=$4 AND unlock=1"
	result, err := db.ExecContext(ctx, query, pb.ExchangeStatus_EXCHANGE_STATUS_CANCEL_BY_USER.Number(),
		exchange.GetId(), exchange.GetUserIdRequest(), pb.ExchangeStatus_EXCHANGE_STATUS_WAITING.Number())
	if err != nil {
		logger.Error("User %s Cancel exchange request id %s, error %s",
			exchange.GetUserIdRequest(), exchange.GetId(), err.Error())
		return nil, status.Error(codes.Internal, "Claim freechip error")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not cancel exchange request.")
		return nil, status.Error(codes.Internal, "Error cancel exchange request")
	}
	exChangeInDb.Status = int64(pb.ExchangeStatus_EXCHANGE_STATUS_CANCEL_BY_USER.Number())
	return exChangeInDb, nil
}

func GetAllExchange(ctx context.Context, logger runtime.Logger, db *sql.DB, userId string, exchange *pb.ExchangeRequest) (*pb.ListExchangeInfo, error) {
	var incomingCursor = &entity.ExchangeListCursor{}
	cusor := exchange.GetCusor()
	if cusor != "" {
		cb, err := base64.URLEncoding.DecodeString(cusor)
		if err != nil {
			return nil, ErrWalletLedgerInvalidCursor
		}
		if err := gob.NewDecoder(bytes.NewReader(cb)).Decode(incomingCursor); err != nil {
			return nil, ErrWalletLedgerInvalidCursor
		}

		// Cursor and filter mismatch. Perhaps the caller has sent an old cursor with a changed filter.
		if userId != incomingCursor.UserId {
			return nil, ErrWalletLedgerInvalidCursor
		}
		logger.Info("GetListFreeChip with cusor id %s, userId %s, create time %s ",
			incomingCursor.Id,
			incomingCursor.UserId,
			incomingCursor.CreateTime.String())
	}
	limit := incomingCursor.Limit
	offset := incomingCursor.Offset

	if limit != exchange.Limit {
		offset = 0
	}
	if limit <= 0 {
		limit = exchange.Limit
	}

	if limit <= 0 {
		limit = 1000
	}

	queryRows := "SELECT id, chips, price, status, unlock, cash_id, cash_type, user_id_request, user_name_request, vip_lv, device_id, user_id_handling, user_name_handling, reason, create_time FROM " +
		ExchangeTableName
	// clause := "WHERE"
	query := ""
	params := make([]interface{}, 0)
	nextClause := func() func(arg string, operator string) string {
		i := 0
		return func(arg string, operator string) string {
			i++
			iStr := "$" + strconv.Itoa(i)
			arg = strings.ToLower(arg)
			if arg == "limit" || arg == "offset" {
				return " " + arg + " " + iStr
			} else {
				if i == 1 {
					return " WHERE " + arg + operator + iStr
				} else {
					return " AND " + arg + operator + iStr
				}
			}

		}
	}()
	if exchange.GetId() != "" {
		if incomingCursor.IsNext {
			query += nextClause("id", "<=")
		} else {
			query += nextClause("id", ">=")
		}
		params = append(params, exchange.GetId())

	}
	if exchange.GetUserIdRequest() != "" {
		query += nextClause("user_id_request", "=")
		params = append(params, exchange.GetUserIdRequest())
	}
	if exchange.GetCashType() != "" {
		query += nextClause("cash_type", "=")
		params = append(params, exchange.GetCashType())
	}

	from := incomingCursor.From
	if from != exchange.GetFrom() {
		offset = 0
	}
	if from <= 0 {
		from = exchange.GetFrom()
	}
	if from > 0 {
		query += nextClause("create_time", ">=")
		params = append(params, time.Unix(from, 0))
	}

	to := incomingCursor.To
	if to != exchange.GetTo() {
		offset = 0
	}
	if to <= 0 {
		to = exchange.GetTo()
	}
	if to > from {
		query += nextClause("create_time", "<=")
		params = append(params, time.Unix(to, 0))
	}

	query += " order by create_time desc "

	query += nextClause("limit", "")
	params = append(params, limit)

	query += nextClause("offset", "")
	params = append(params, offset)

	var dbChips, dbStatus, dbVipLv int64
	var dbId, dbPrice, dbCashId, dbCashType, dbUserIdReq, dbUserNameReq,
		dbDeviceId, dbUserIdHandling, dbUserNameHandling, dbReason string
	var dbUnlock int32
	var dbCreateTime pgtype.Timestamptz
	// logger.Debug("Query %s", query)
	rows, err := db.QueryContext(ctx, queryRows+query, params...)

	if err != nil {
		logger.Error("Query exchange id %s, error %s", exchange.Id, err.Error())
		return nil, status.Error(codes.Internal, "Query exchange error")
	}
	//  id, chips, price,
	//  status, unlock, cash_id,
	//   cash_type, user_id_request, user_name_request,
	//   vip_lv, device_id, user_id_handling,
	//   user_name_handling, reason, create_time FROM
	ml := make([]*pb.ExchangeInfo, 0)
	for rows.Next() {
		rows.Scan(&dbId, &dbChips, &dbPrice,
			&dbStatus, &dbUnlock, &dbCashId,
			&dbCashType, &dbUserIdReq, &dbUserNameReq,
			&dbVipLv, &dbDeviceId, &dbUserIdHandling,
			&dbUserNameHandling, &dbReason, &dbCreateTime)
		exchangeInfo := pb.ExchangeInfo{
			Id:               dbId,
			Chips:            dbChips,
			Price:            dbPrice,
			Status:           dbStatus,
			Unlock:           dbUnlock,
			CashId:           dbCashId,
			CashType:         dbCashType,
			UserIdRequest:    dbUserIdReq,
			UserNameRequest:  dbUserNameReq,
			VipLv:            dbVipLv,
			DeviceId:         dbDeviceId,
			UserIdHandling:   dbUserIdHandling,
			UserNameHandling: dbUserNameHandling,
			Reason:           dbReason,
			CreateTime:       dbCreateTime.Time.Unix(),
		}
		ml = append(ml, &exchangeInfo)
	}
	sort.Slice(ml, func(i, j int) bool {
		return ml[i].CreateTime > ml[j].CreateTime
	})

	var total int64 = incomingCursor.Total
	if total <= 0 {
		queryTotal := "Select count(*) as total FROM " + ExchangeTableName + " " +
			strings.ReplaceAll(query, "order by create_time desc", "")
		// logger.Debug("Query total %s", queryTotal)
		e := db.QueryRowContext(ctx, queryTotal, params...).Scan(&total)
		if e != nil {
			logger.Error(e.Error())
		}
	}

	var nextCursor *entity.ExchangeListCursor
	var prevCursor *entity.ExchangeListCursor
	if len(ml) > 0 {
		if len(ml)+int(incomingCursor.Offset) < int(total) {
			nextCursor = &entity.ExchangeListCursor{
				UserId: userId,
				Id:     ml[len(ml)-1].Id,
				IsNext: true,
				Offset: offset + int64(len(ml)),
				Limit:  limit,
				From:   from,
				To:     to,
				Total:  total,
			}
		}

		prevOffset := incomingCursor.Offset - int64(len(ml))
		if len(ml)+int(incomingCursor.Offset) >= int(total) {
			prevOffset = total - int64(len(ml)) - limit
		}
		if prevOffset < 0 {
			prevOffset = 0
		}
		if incomingCursor.Offset > 0 {
			prevCursor = &entity.ExchangeListCursor{
				UserId: userId,
				Id:     ml[0].Id,
				IsNext: false,
				Offset: prevOffset,
				Total:  total,
				Limit:  limit,
				From:   from,
				To:     to,
			}
		}

	}

	var nextCursorStr string
	if nextCursor != nil {
		cursorBuf := new(bytes.Buffer)
		if err := gob.NewEncoder(cursorBuf).Encode(nextCursor); err != nil {
			logger.Error("Error creating wallet ledger list cursor", zap.Error(err))
			return nil, err
		}
		nextCursorStr = base64.URLEncoding.EncodeToString(cursorBuf.Bytes())
	}

	var prevCursorStr string
	if prevCursor != nil {
		cursorBuf := new(bytes.Buffer)
		if err := gob.NewEncoder(cursorBuf).Encode(prevCursor); err != nil {
			logger.Error("Error creating wallet ledger list cursor", zap.Error(err))
			return nil, err
		}
		prevCursorStr = base64.URLEncoding.EncodeToString(cursorBuf.Bytes())
	}

	return &pb.ListExchangeInfo{
		ExchangeInfos: ml,
		NextCusor:     nextCursorStr,
		PrevCusor:     prevCursorStr,
		Total:         total,
		Offset:        incomingCursor.Offset,
		Limit:         limit,
		From:          from,
		To:            to,
	}, nil
}

func ExchangeLock(ctx context.Context, logger runtime.Logger, db *sql.DB, exchange *pb.ExchangeInfo) (*pb.ExchangeInfo, error) {
	curExchange, err := GetExchangeById(ctx, logger, db, exchange)
	if err != nil {
		logger.Error("get exchange id %s error %s", exchange.GetId(), err.Error())
		return nil, status.Error(codes.Internal, "get exchange error")
	}
	if curExchange.Unlock == 0 || curExchange.Status != int64(pb.ExchangeStatus_EXCHANGE_STATUS_WAITING.Number()) {
		return curExchange, nil
	}
	query := "UPDATE " + ExchangeTableName + " SET status=$1, unlock=0, update_time = now() WHERE id=$2 AND status=$3"
	result, err := db.ExecContext(ctx, query,
		pb.ExchangeStatus_EXCHANGE_STATUS_PENDING.Number(),
		exchange.GetId(),
		pb.ExchangeStatus_EXCHANGE_STATUS_WAITING.Number())
	if err != nil {
		logger.Error("Lock exchange id %s error %s", exchange.GetId(), err.Error())
		return nil, status.Error(codes.Internal, "Lock exchange error")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not lock exchange %s", exchange.GetId())
		return nil, status.Error(codes.Internal, "Error lock exchange")
	}
	return GetExchangeById(ctx, logger, db, exchange)
}

func ExchangeUpdateStatus(ctx context.Context, logger runtime.Logger, db *sql.DB, exchange *pb.ExchangeInfo) (*pb.ExchangeInfo, error) {
	curExchange, err := GetExchangeById(ctx, logger, db, exchange)
	if err != nil {
		logger.Error("get exchange id %s error %s", exchange.GetId(), err.Error())
		return nil, status.Error(codes.Internal, "get exchange error")
	}
	if curExchange.Unlock != 0 ||
		curExchange.GetStatus() != int64(pb.ExchangeStatus_EXCHANGE_STATUS_PENDING.Number()) ||
		(exchange.GetStatus() != int64(pb.ExchangeStatus_EXCHANGE_STATUS_REJECT.Number()) &&
			exchange.GetStatus() != int64(pb.ExchangeStatus_EXCHANGE_STATUS_DONE.Number())) {
		logger.Error("Can not update status exchange. Not meet requirement,", curExchange.Unlock)
		return curExchange, errors.New("can not update status exchange. Not meet requirement")
	}
	query := "UPDATE " + ExchangeTableName + " SET status=$1, reason=$2, update_time = now() WHERE id=$3 AND status=$4"
	result, err := db.ExecContext(ctx, query,
		exchange.Status,
		exchange.Reason,
		exchange.GetId(),
		pb.ExchangeStatus_EXCHANGE_STATUS_PENDING.Number())
	if err != nil {
		logger.Error("Update status exchange id %s error %s", exchange.GetId(), err.Error())
		return nil, status.Error(codes.Internal, "Lock exchange error")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not update status exchange %s", exchange.GetId())
		return nil, status.Error(codes.Internal, "Error update status exchange")
	}
	return GetExchangeById(ctx, logger, db, exchange)
}
func TotalCashoutByUsers(ctx context.Context, db *sql.DB, userIds ...string) ([]*pb.CashOut, error) {
	query := `SELECT user_id_request, coalesce(sum(chips),0) as chips
FROM public.exchange where user_id_request IN (` + "'" + strings.Join(userIds, "','") + "'" + `) group by user_id_request;
`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	var userId string
	var chips int64
	ml := make([]*pb.CashOut, 0)
	for rows.Next() {
		err = rows.Scan(&userId, &chips)
		if err != nil {
			continue
		}
		v := &pb.CashOut{
			UserId: userId,
			Co:     chips,
		}
		ml = append(ml, v)
	}
	return ml, nil
}
func TotalCashoutInTimeByUsers(ctx context.Context, db *sql.DB, fromUnix, toUnix int64, userIds ...string) ([]*pb.CashOut, error) {
	query := `SELECT user_id_request, coalesce(sum(chips), 0) as chips
FROM public.exchange where  create_time >=$1 and create_time <=$2   and 
	user_id_request IN (` + "'" + strings.Join(userIds, "','") + "'" + `) group by user_id_request;`
	rows, err := db.QueryContext(ctx, query,
		time.Unix(fromUnix, 0), time.Unix(toUnix, 0))
	if err != nil {
		return nil, err
	}
	var userId string
	var chips int64
	ml := make([]*pb.CashOut, 0)
	for rows.Next() {
		err = rows.Scan(&userId, &chips)
		if err != nil {
			continue
		}
		v := &pb.CashOut{
			Coo: chips,
		}
		ml = append(ml, v)
	}
	return ml, nil
}

func FilterUsersByTotalCashout(ctx context.Context, db *sql.DB, condition string, value int64) ([]*pb.CashOut, error) {
	query := `SELECT user_id_request, coalesce(sum(chips), 0) as chips
FROM public.exchange group by user_id_request having coalesce(sum(chips), 0) ` + condition + ` $1;`
	rows, err := db.QueryContext(ctx, query, value)
	if err != nil {
		return nil, err
	}
	var userId string
	var chips int64
	ml := make([]*pb.CashOut, 0)
	for rows.Next() {
		err = rows.Scan(&userId, &chips)
		if err != nil {
			continue
		}
		v := &pb.CashOut{
			Coo:    chips,
			UserId: userId,
		}
		ml = append(ml, v)
	}
	return ml, nil
}

func FilterUsersByTotalCashoutInTime(ctx context.Context, db *sql.DB, fromUnix, toUnix int64, condition string, value int64) ([]*pb.CashOut, error) {
	query := `SELECT user_id_request, coalesce(sum(chips), 0) as chips
FROM public.exchange where  create_time >=$1 and create_time <=$2 group by user_id_request having coalesce(sum(chips), 0) ` + condition + ` $3;`
	rows, err := db.QueryContext(ctx, query,
		time.Unix(fromUnix, 0), time.Unix(toUnix, 0), value)
	if err != nil {
		return nil, err
	}
	var userId string
	var chips int64
	ml := make([]*pb.CashOut, 0)
	for rows.Next() {
		err = rows.Scan(&userId, &chips)
		if err != nil {
			continue
		}
		v := &pb.CashOut{
			Coo:    chips,
			UserId: userId,
		}
		ml = append(ml, v)
	}
	return ml, nil
}
