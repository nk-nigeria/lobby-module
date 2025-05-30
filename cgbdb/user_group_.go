package cgbdb

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakama-nigeria/lobby-module/entity"
	pb "github.com/nakama-nigeria/cgp-common/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"sort"
	"strings"
	"time"
)

func getCashOutData(ctx context.Context, logger runtime.Logger, db *sql.DB, userId string, data *entity.UserGroupUserInfo) {
	cashOut, err := TotalCashoutByUsers(ctx, db, userId)
	if err != nil || len(cashOut) == 0 {
		logger.Error("TotalCashoutByUsers %v %d", err, len(cashOut))
	} else {
		data.Co = cashOut[0].Co
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 99, now.Location())
	cashOut, err = TotalCashoutInTimeByUsers(ctx, db, start.Unix(), end.Unix(), userId)
	if err != nil || len(cashOut) == 0 {
		logger.Error("TotalCashoutByUsers %v %d", err, len(cashOut))
	} else {
		data.CO0 = cashOut[0].Co
	}
}

func getCashInData(ctx context.Context, logger runtime.Logger, db *sql.DB, userId string, data *entity.UserGroupUserInfo) {
	cashIn, err := TotalDepositByUsers(ctx, db, userId)
	if err != nil || len(cashIn) == 0 {
		logger.Error("TotalDepositByUsers %v %d", err, len(cashIn))
	} else {
		data.LQ = cashIn[0].Ci
	}

	// cashin by range time
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 99, now.Location())
	cashIn, err = TotalDepositInTimeByUsers(ctx, db, start.Unix(), end.Unix(), userId)
	if err != nil || len(cashIn) == 0 {
		logger.Error("TotalDepositInTimeByUsers %v %d", err, len(cashIn))
	} else {
		data.BLQ1 = cashIn[0].Ci
	}

	start = time.Date(now.Year(), now.Month(), now.Day()-2, 0, 0, 0, 0, now.Location())
	cashIn, err = TotalDepositInTimeByUsers(ctx, db, start.Unix(), end.Unix(), userId)
	if err != nil || len(cashIn) == 0 {
		logger.Error("TotalDepositInTimeByUsers %v %d", err, len(cashIn))
	} else {
		data.BLQ3 = cashIn[0].Ci
	}

	start = time.Date(now.Year(), now.Month(), now.Day()-4, 0, 0, 0, 0, now.Location())
	cashIn, err = TotalDepositInTimeByUsers(ctx, db, start.Unix(), end.Unix(), userId)
	if err != nil || len(cashIn) == 0 {
		logger.Error("TotalDepositInTimeByUsers %v %d", err, len(cashIn))
	} else {
		data.BLQ5 = cashIn[0].Ci
	}

	start = time.Date(now.Year(), now.Month(), now.Day()-6, 0, 0, 0, 0, now.Location())
	cashIn, err = TotalDepositInTimeByUsers(ctx, db, start.Unix(), end.Unix(), userId)
	if err != nil || len(cashIn) == 0 {
		logger.Error("TotalDepositInTimeByUsers %v %d", err, len(cashIn))
	} else {
		data.BLQ7 = cashIn[0].Ci
	}

	start = time.Date(now.Year(), now.Month(), now.Day()-6, 0, 0, 0, 0, now.Location())
	cashIn, err = AvgDepositInTimeByUsers(ctx, db, start.Unix(), end.Unix(), userId)
	if err != nil || len(cashIn) == 0 {
		logger.Error("TotalDepositInTimeByUsers %v %d", err, len(cashIn))
	} else {
		data.Avgtrans7 = cashIn[0].Ci
	}
}

func GetUserGroupUserInfo(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, userId string) (*entity.UserGroupUserInfo, error) {
	result := &entity.UserGroupUserInfo{}

	account, err := nk.AccountGetId(ctx, userId)
	if err != nil {
		logger.Error("GetAccount error %s", err.Error())
		return result, err
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(account.User.GetMetadata()), &metadata); err != nil {
		return result, errors.New("Corrupted user metadata.")
	}
	var level int64 = 0
	if levelStr, ok := metadata["level"]; ok {
		level = levelStr.(int64)
	}
	result.Level = level

	result.VipLevel = entity.ToInt64(metadata["vip_level"], 0)

	var chips int64 = 0
	var chipsInbank int64 = 0
	wallet, err := entity.ParseWallet(account.Wallet)
	if err == nil {
		chips = wallet.Chips
		chipsInbank = wallet.ChipsInBank
	}
	result.ChipsInBank = chipsInbank
	result.AG = chips

	getCashOutData(ctx, logger, db, userId, result)
	getCashInData(ctx, logger, db, userId, result)

	if account.User.GetCreateTime() != nil {
		result.CreateTime = account.User.GetCreateTime().GetSeconds()
	}
	return result, nil
}

func GetListUserIdsByUserGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, unmarshaler *protojson.UnmarshalOptions, id int64) ([]string, error) {
	userGroup, err := GetUserGroupById(ctx, logger, db, unmarshaler, id)
	if err != nil || userGroup.Name == "" {
		return nil, err
	}
	condition := ""
	params := make([]interface{}, 0)
	logger.Info("FetchUserIDWithCondition %v %v", condition, params)
	return FetchUserIDWithCondition(ctx, db, condition, params)
}

func GetListUserGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, unmarshaler *protojson.UnmarshalOptions, limit int64, cursor string) (*pb.ListUserGroup, error) {
	var incomingCursor = &entity.UserGroupListCursor{}
	if cursor != "" {
		cb, err := base64.URLEncoding.DecodeString(cursor)
		if err != nil {
			return nil, ErrWalletLedgerInvalidCursor
		}
		if err := gob.NewDecoder(bytes.NewReader(cb)).Decode(incomingCursor); err != nil {
			return nil, ErrWalletLedgerInvalidCursor
		}

		logger.Info("GetListUserGroup with cusor %d, create time %s ",
			incomingCursor.Id,
			incomingCursor.CreateTime.String())
	}

	if limit <= 0 {
		limit = 100
	}
	if incomingCursor.Id < 0 {
		incomingCursor.Id = 0
	}

	var rows *sql.Rows
	var err error

	params := make([]interface{}, 0)
	query := ""

	if incomingCursor.Id > 0 {
		if incomingCursor.IsNext {
			query += " WHERE id < $1 AND deleted = false order by id desc "
		} else {
			query += " WHERE id > $1 AND deleted = false order by id asc"
		}
		params = append(params, incomingCursor.Id)
		query += "  limit $2"
		params = append(params, limit)
	} else {
		query += " WHERE deleted = false  order by id desc limit $1"
		params = append(params, limit)
	}
	queryRow := "SELECT id, name, type, condition FROM " +
		UserGroupTableName + query
	rows, err = db.QueryContext(ctx, queryRow, params...)
	if err != nil {
		logger.Error("Query lists user group, error %s", err.Error())
		return nil, status.Error(codes.Internal, "Query lists user group")
	}
	ml := make([]*pb.UserGroup, 0)
	var dbID int64
	var dbName, dbType string
	var dbCondition []byte
	for rows.Next() {
		rows.Scan(&dbID, &dbName, &dbType, &dbCondition)
		userGroup := pb.UserGroup{
			Id:   dbID,
			Name: dbName,
		}
		ml = append(ml, &userGroup)
	}
	sort.Slice(ml, func(i, j int) bool {
		return ml[i].Id > ml[j].Id
	})
	var total int64 = incomingCursor.Total
	if total <= 0 {
		queryTotal := "Select count(*) as total FROM " + UserGroupTableName +
			strings.ReplaceAll(query, "order by id desc", "")

		_ = db.QueryRowContext(ctx, queryTotal, params...).Scan(&total)
	}
	var nextCursor *entity.UserGroupListCursor
	var prevCursor *entity.UserGroupListCursor
	if len(ml) > 0 {
		if len(ml)+int(incomingCursor.Offset) < int(total) {
			nextCursor = &entity.UserGroupListCursor{
				Id:     ml[len(ml)-1].Id,
				IsNext: true,
				Offset: incomingCursor.Offset + int64(len(ml)),
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
		prevCursor = &entity.UserGroupListCursor{
			Id:     ml[0].Id,
			IsNext: false,
			Offset: prevOffset,
			Total:  total,
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

	return &pb.ListUserGroup{
		UserGroups: ml,
		NextCusor:  nextCursorStr,
		PrevCusor:  prevCursorStr,
		Total:      total,
		Offset:     incomingCursor.Offset,
		Limit:      limit,
	}, nil
}
