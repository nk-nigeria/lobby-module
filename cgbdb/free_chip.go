package cgbdb

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"sort"
	"strings"

	"github.com/ciaolink-game-platform/cgb-lobby-module/conf"
	"github.com/ciaolink-game-platform/cgb-lobby-module/entity"
	pb "github.com/ciaolink-game-platform/cgp-common/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CREATE TABLE public.freechip (
//
//	id bigint NOT NULL,
//	sender_id character varying(128) NOT NULL,
//	recipient_id character varying(128) NOT NULL,
//	title character varying(128) NOT NULL,
//	content character varying(128) NOT NULL,
//	chips integer NOT NULL DEFAULT 0,
//	claimable smallint NOT NULL DEFAULT 1,
//	action character varying(128) NOT NULL,
//	create_time timestamp with time zone NOT NULL DEFAULT now(),
//	update_time timestamp with time zone NOT NULL DEFAULT now()
//
// );
// ALTER TABLE
//
//	public.freechip
//
// ADD
//
//	CONSTRAINT freechip_pkey PRIMARY KEY (id)
const FreeChipTableName = "freechip"

func AddClaimableFreeChip(ctx context.Context, logger runtime.Logger, db *sql.DB, freeChip *pb.FreeChip) error {
	if freeChip == nil || freeChip.RecipientId == "" || freeChip.Chips <= 0 {
		return status.Error(codes.InvalidArgument, "Error add claimable freechip.")
	}
	freeChip.Id = conf.SnowlakeNode.Generate().Int64()
	query := "INSERT INTO " + FreeChipTableName + " (id, sender_id, recipient_id, title, content, chips, claimable, action, create_time, update_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())"
	result, err := db.ExecContext(ctx, query, freeChip.Id, freeChip.SenderId, freeChip.RecipientId, freeChip.Title, freeChip.Content,
		freeChip.Chips, 1, freeChip.GetAction())
	if err != nil {
		logger.Error("Add new claimable, sender: %s, recv: %s, chips: %d, error %s",
			freeChip.SenderId, freeChip.RecipientId, freeChip.Chips, err.Error())
		return status.Error(codes.Internal, "Error add claimable freechip.")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert new claimable, sender: %s, recv: %s, chips: %s",
			freeChip.SenderId, freeChip.RecipientId, freeChip.Chips)
		return status.Error(codes.Internal, "Error add claimable freechip.")
	}
	return nil
}

func ClaimFreeChip(ctx context.Context, logger runtime.Logger, db *sql.DB, id int64, recipientId string) (*pb.FreeChip, error) {
	freeChip, err := GetFreeChipByIdByUser(ctx, logger, db, id, recipientId)
	if err != nil {
		return nil, err
	}
	if !freeChip.Claimable {
		return nil, status.Error(codes.Aborted, "Freechip alread claimed")
	}
	query := "UPDATE " + FreeChipTableName + " SET claimable=$1 WHERE id=$2 AND recipient_id=$3 AND claimable=$4"
	result, err := db.ExecContext(ctx, query, 0, id, recipientId, 1)
	if err != nil {
		logger.Error("Claim free chip id %d, user %s, error %s", id, recipientId, err.Error())
		return nil, status.Error(codes.Internal, "Claim freechip error")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not claim freechip.")
		return nil, status.Error(codes.Internal, "Error claim freechip")
	}
	return freeChip, nil
}

func GetFreeChipByIdByUser(ctx context.Context, logger runtime.Logger, db *sql.DB, id int64, recipientId string) (*pb.FreeChip, error) {
	if id <= 0 || recipientId == "" {
		return nil, status.Error(codes.InvalidArgument, "Id or user id is empty")
	}
	query := "SELECT id, sender_id, recipient_id, title, content, chips, claimable, action FROM " + FreeChipTableName + " WHERE id=$1 AND recipient_id=$2"
	var dbID int64
	var dbSenderId, dbRecvId, dbTitle, dbContent, dbAction string
	var dbChips int64
	var dbClaimable int
	err := db.QueryRowContext(ctx, query, id, recipientId).Scan(&dbID, &dbSenderId, &dbRecvId, &dbTitle, &dbContent, &dbChips, &dbClaimable, &dbAction)
	if err != nil {
		logger.Error("Query free chip id %, user %s, error %s", id, recipientId, err.Error())
		return nil, status.Error(codes.Internal, "Query freechip error")
	}
	freeChip := pb.FreeChip{
		Id:          dbID,
		SenderId:    dbSenderId,
		RecipientId: dbRecvId,
		Title:       dbTitle,
		Content:     dbContent,
		Chips:       dbChips,
		Action:      dbAction,
	}
	if dbClaimable == 1 {
		freeChip.Claimable = true
	}
	return &freeChip, nil
}

func GetFreeChipClaimableByUser(ctx context.Context, logger runtime.Logger, db *sql.DB, recipientId string) (*pb.ListFreeChip, error) {
	if recipientId == "" {
		return nil, status.Error(codes.InvalidArgument, "Id or user id is empty")
	}
	query := "SELECT id, sender_id, recipient_id, title, content, chips, claimable, action FROM " + FreeChipTableName + " WHERE claimable=$1 AND recipient_id=$2"

	rows, err := db.QueryContext(ctx, query, 1, recipientId)
	if err != nil {
		logger.Error("Query free chip claimable user %s, error %s", recipientId, err.Error())
		return nil, status.Error(codes.Internal, "Query freechip claimable error")
	}
	ml := make([]*pb.FreeChip, 0)
	var dbID int64
	var dbSenderId, dbRecvId, dbTitle, dbContent, dbAction string
	var dbChips int64
	var dbClaimable int
	for rows.Next() {
		rows.Scan(&dbID, &dbSenderId, &dbRecvId, &dbTitle, &dbContent, &dbChips, &dbClaimable, &dbAction)
		freeChip := pb.FreeChip{
			Id:          dbID,
			SenderId:    dbSenderId,
			RecipientId: dbRecvId,
			Title:       dbTitle,
			Content:     dbContent,
			Chips:       dbChips,
			Action:      dbAction,
		}
		if dbClaimable == 1 {
			freeChip.Claimable = true
		}
		ml = append(ml, &freeChip)
	}

	return &pb.ListFreeChip{
		Freechips: ml,
	}, nil
}

func GetListFreeChip(ctx context.Context, logger runtime.Logger, db *sql.DB, recipientId string, limit int64, cursor string) (*pb.ListFreeChip, error) {
	var incomingCursor = &entity.FreeChipListCursor{}
	if cursor != "" {
		cb, err := base64.URLEncoding.DecodeString(cursor)
		if err != nil {
			return nil, ErrWalletLedgerInvalidCursor
		}
		if err := gob.NewDecoder(bytes.NewReader(cb)).Decode(incomingCursor); err != nil {
			return nil, ErrWalletLedgerInvalidCursor
		}

		// Cursor and filter mismatch. Perhaps the caller has sent an old cursor with a changed filter.
		if recipientId != incomingCursor.UserId {
			return nil, ErrWalletLedgerInvalidCursor
		}
		logger.Info("GetListFreeChip with cusor %d, userId %s, create time %s ",
			incomingCursor.Id,
			incomingCursor.UserId,
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
	// if userId == "" {

	// 	if incomingCursor != nil && !incomingCursor.IsNext {
	// 		params = append(params, incomingCursor.Id)
	// 		query = "SELECT id, sender_id, recipient_id, title, content, chips, claimable FROM " + FreeChipTableName + " WHERE id < $1 order by id desc limit $2"
	// 	} else {
	// 		params = append(params, 0)
	// 	}
	// 	logger.Info("query %s", query)
	// 	params = append(params, limit)
	// 	rows, err = db.QueryContext(ctx, query, params...)
	// } else {
	// 	params := make([]interface{}, 0)
	// 	query := "SELECT id, sender_id, recipient_id, title, content, chips, claimable FROM " + FreeChipTableName + " WHERE recipient_id=$1 AND id > $2 order by id desc limit $3"
	// 	params = append(params, userId)
	// 	if incomingCursor != nil && !incomingCursor.IsNext {
	// 		query = "SELECT id, sender_id, recipient_id, title, content, chips, claimable FROM " + FreeChipTableName + " WHERE recipient_id=$1 AND id < $2 order by id desc limit $3"
	// 		params = append(params, incomingCursor.Id)
	// 	} else {
	// 		params = append(params, 0)
	// 	}
	// 	params = append(params, limit)
	// 	logger.Info("query %s", query)
	// 	rows, err = db.QueryContext(ctx, query, params...)
	// }
	params := make([]interface{}, 0)
	query := ""

	if recipientId == "" {
		if incomingCursor.Id > 0 {
			if incomingCursor.IsNext {
				query += " WHERE id < $1 order by id desc "
			} else {
				query += " WHERE id > $1 order by id asc"
			}
			params = append(params, incomingCursor.Id)
			query += "  limit $2"
			params = append(params, limit)
		} else {
			query += " order by id desc limit $1"
			params = append(params, limit)
		}
	} else {
		query += " WHERE recipient_id=$1 "
		params = append(params, recipientId)
		if incomingCursor.Id > 0 {
			if incomingCursor.IsNext {
				query += " AND id < $2 order by id desc "
			} else {
				query += " AND id > $2 order by id asc "
			}
			params = append(params, incomingCursor.Id)
			query += " limit $3"
			params = append(params, limit)
		} else {
			query += " order by id desc limit $2"
			params = append(params, limit)
		}
	}
	queryRow := "SELECT id, sender_id, recipient_id, title, content, chips, claimable, action FROM " +
		FreeChipTableName + query
	rows, err = db.QueryContext(ctx, queryRow, params...)
	if err != nil {
		logger.Error("Query lists free chip claimable user %s, error %s", recipientId, err.Error())
		return nil, status.Error(codes.Internal, "Query freechip claimable error")
	}
	ml := make([]*pb.FreeChip, 0)
	var dbID int64
	var dbSenderId, dbRecvId, dbTitle, dbContent, dbAction string
	var dbChips int64
	var dbClaimable int
	for rows.Next() {
		rows.Scan(&dbID, &dbSenderId, &dbRecvId, &dbTitle, &dbContent, &dbChips, &dbClaimable, &dbAction)
		freeChip := pb.FreeChip{
			Id:          dbID,
			SenderId:    dbSenderId,
			RecipientId: dbRecvId,
			Title:       dbTitle,
			Content:     dbContent,
			Chips:       dbChips,
			Action:      dbAction,
		}
		if dbClaimable == 1 {
			freeChip.Claimable = true
		}
		ml = append(ml, &freeChip)
	}
	sort.Slice(ml, func(i, j int) bool {
		return ml[i].Id > ml[j].Id
	})
	var total int64 = incomingCursor.Total
	if total <= 0 {
		queryTotal := "Select count(*) as total FROM " + FreeChipTableName +
			strings.ReplaceAll(query, "order by id desc", "")

		_ = db.QueryRowContext(ctx, queryTotal, params...).Scan(&total)
	}
	var nextCursor *entity.FreeChipListCursor
	var prevCursor *entity.FreeChipListCursor
	if len(ml) > 0 {
		if len(ml)+int(incomingCursor.Offset) < int(total) {
			nextCursor = &entity.FreeChipListCursor{
				UserId: recipientId,
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
		prevCursor = &entity.FreeChipListCursor{
			UserId: recipientId,
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

	return &pb.ListFreeChip{
		Freechips: ml,
		NextCusor: nextCursorStr,
		PrevCusor: prevCursorStr,
		Total:     total,
		Offset:    incomingCursor.Offset,
		Limit:     limit,
	}, nil
}
