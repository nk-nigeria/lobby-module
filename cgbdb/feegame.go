package cgbdb

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const FeeGameTableName = "feegame"

// CREATE TABLE
//   public.feegame (
//     id bigint NOT NULL,
// 	   user_id character varying(128) NOT NULL,
// 	   game character varying(128) NOT NULL,
//     fee bigint NOT NULL DEFAULT 0,
//     create_time timestamp
//     with
//       time zone NOT NULL DEFAULT now(),
//       update_time timestamp
//     with
//       time zone NOT NULL DEFAULT now()
//   );

// ALTER TABLE
//   public.feegame
// ADD
//   CONSTRAINT feegame_pkey PRIMARY KEY (id)
func GetListFeeGame(ctx context.Context, logger runtime.Logger, db *sql.DB, req *entity.FeeGameListCursor) ([]entity.FeeGame, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing user id")
	}
	query := "SELECT id, user_id, game, fee,create_time FROM " + FeeGameTableName +
		" WHERE user_id=$1"
	args := make([]interface{}, 0)
	idx := 1
	args = append(args, req.UserId)
	if req.From > 0 {
		idx++
		query += " AND create_time>=$" + strconv.Itoa(idx)
		args = append(args, time.Unix(req.From, 0))

	}
	if req.To > 0 {
		idx++
		query += " AND create_time <=$" + strconv.Itoa(idx)
		args = append(args, time.Unix(req.To, 0))

	}
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Error("Query list fee game user %s, err %s", req.UserId, err.Error())
		return nil, status.Error(codes.Internal, "Query list fee game error")
	}
	var dbId int64
	var dbUserId, dbGame string
	var dbFee int64
	var dbCreateTime pgtype.Timestamptz
	ml := make([]entity.FeeGame, 0)
	for rows.Next() {
		if rows.Scan(&dbId, &dbUserId, &dbGame, &dbFee, &dbCreateTime) == nil {
			ml = append(ml, entity.FeeGame{
				Id:             dbId,
				UserID:         dbUserId,
				Game:           dbGame,
				Fee:            dbFee,
				CreateTimeUnix: dbCreateTime.Time.Unix(),
				From:           req.From,
				To:             req.To,
			})
		}
	}
	return ml, nil
}

func GetSumFeeByUserId(ctx context.Context, logger runtime.Logger, db *sql.DB, req *entity.FeeGameListCursor) (entity.FeeGame, error) {
	if req.UserId == "" {
		return entity.FeeGame{}, status.Error(codes.InvalidArgument, "Missing user id")
	}
	query := "SELECT sum(fee) from " + FeeGameTableName + " where user_id=$1"
	args := make([]interface{}, 0)
	args = append(args, req.UserId)
	idx := 1
	if req.From > 0 {
		idx++
		query += " AND create_time>=$" + strconv.Itoa(idx)

		args = append(args, time.Unix(req.From, 0))
	}
	if req.To > 0 {
		idx++
		query += " AND create_time <=$" + strconv.Itoa(idx)
		args = append(args, time.Unix(req.To, 0))
	}
	var dbSumFree sql.NullInt64
	err := db.QueryRowContext(ctx, query, args...).Scan(&dbSumFree)
	if err != nil {
		logger.Error("Get sum fee game user %s, error %s", req.UserId, err.Error())
		return entity.FeeGame{}, status.Error(codes.Internal, "get sum free game error")
	}
	l := entity.FeeGame{
		UserID: req.UserId,
		From:   req.From,
		To:     req.To,
	}
	// check if sum(fee) is null
	if dbSumFree.Valid {
		l.Fee = dbSumFree.Int64
	}
	return l, nil
}
