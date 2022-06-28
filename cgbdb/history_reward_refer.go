package cgbdb

import (
	"context"
	"database/sql"

	"github.com/ciaolink-game-platform/cgp-lobby-module/conf"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CREATE TABLE
// public.history_reward_refer (
// 	id bigint NOT NULL,
// 	user_id character varying(128) NOT NULL,
// 	win_amt bigint NOT NULL,
// 	reward bigint NOT NULL,
// // 	data VARCHAR,
// from_unix bigint ,
// 				to_unix bigint,
// 	create_time timestamp
// with
// time zone NOT NULL DEFAULT now(),
// update_time timestamp
// with
// time zone NOT NULL DEFAULT now()
// );

// ALTER TABLE
// public.history_reward_refer
// ADD
// CONSTRAINT history_reward_refer_pkey PRIMARY KEY (id)

const HistoryRewardReferTableName = "history_reward_refer"

func AddNewHistoryRewardRefer(ctx context.Context, logger runtime.Logger, db *sql.DB, reward *pb.RewardRefer) (int64, error) {
	dbId := conf.SnowlakeNode.Generate().Int64()
	query := "INSERT INTO " + HistoryRewardReferTableName +
		" (id, user_id, win_amt, reward, data, from_unix, to_unix, create_time, update_time) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())"
	l := pb.ListRewardRefer{
		UserRefers: reward.UserRefers,
	}
	data, err := conf.Marshaler.Marshal(&l)
	if err != nil {
		logger.Error("Marshalerd data error %s", err.Error())
		return 0, err
	}
	result, err := db.ExecContext(ctx, query,
		dbId, reward.GetUserId(), reward.GetWinAmt(), reward.GetEstReward(), data,
		reward.GetFrom(), reward.GetTo())
	if err != nil {
		logger.Error("Error when add history reward refer user, user : %s, error %s",
			reward.UserId, err.Error())
		return 0, status.Error(codes.Internal, "Error add history reward refer user.")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert add history reward refer user, user : %s",
			reward.UserId)
		return 0, status.Error(codes.Internal, "Error add history reward refer user.")
	}
	return dbId, nil
}

func GetHistoryRewardRefer(ctx context.Context, logger runtime.Logger, db *sql.DB, req *pb.HistoryRewardRequest) ([]*pb.RewardRefer, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing user id")
	}
	if req.GetFrom() <= 0 || req.GetTo() <= 0 || req.GetFrom() >= req.GetTo() {
		return nil, status.Error(codes.InvalidArgument, "Invalid time")
	}
	query := "SELECT id, user_id, win_amt, reward, data, from_unix, to_unix FROM " +
		HistoryRewardReferTableName + " WHERE user_id=$1 AND from_unix >= $2 AND to_unix <= $3"
	rows, err := db.QueryContext(ctx, query,
		req.GetUserId(), req.From, req.To)
	if err != nil {
		logger.Error("Query history reward refer user %s error %s", req.GetUserId(), err.Error())
		return nil, status.Error(codes.Internal, "Query history reward refer error")
	}
	ml := make([]*pb.RewardRefer, 0)
	var dbID int64
	var dbUserId, dbData string
	var dbWinAmt, dbReward int64
	var dbFrom, dbTo int64
	for rows.Next() {
		if rows.Scan(&dbID, &dbUserId, &dbWinAmt, &dbReward, &dbData, &dbFrom, &dbTo) == nil {
			r := &pb.RewardRefer{
				UserId:    dbUserId,
				WinAmt:    dbWinAmt,
				EstReward: dbReward,
			}
			l := pb.ListRewardRefer{}
			if conf.Unmarshaler.Unmarshal([]byte(dbData), &l) == nil {
				r.UserRefers = l.GetUserRefers()
			}
			ml = append(ml, r)
		}
	}
	return ml, nil
}
