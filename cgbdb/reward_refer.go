package cgbdb

import (
	"context"
	"database/sql"
	"strings"

	"github.com/ciaolink-game-platform/cgp-lobby-module/conf"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CREATE TABLE
// public.reward_refer (
// 	id bigint NOT NULL,
// 	user_id character varying(128) NOT NULL,
// 	win_amt bigint NOT NULL,
// 	reward bigint NOT NULL,
// 	reward_lv integer NOT NULL,
// 	reward_rate double precision NOT NULL DEFAULT 0,
// 	data VARCHAR,
//  send_to_wallet smallint NOT NULL DEFAULT 0,
//  from_unix bigint ,
// 	to_unix bigint,
// UNIQUE (user_id, from_unix, to_unix),
// 	create_time timestamp
// with
// time zone NOT NULL DEFAULT now(),
// update_time timestamp
// with
// time zone NOT NULL DEFAULT now()
// );

// ALTER TABLE
// public.reward_refer
// ADD
// CONSTRAINT reward_refer_pkey PRIMARY KEY (id)

const RewardReferTableName = "reward_refer"

func AddOrUpdateIfExistRewardRefer(ctx context.Context, logger runtime.Logger, db *sql.DB, reward *pb.RewardRefer) (int64, error) {
	dbId := conf.SnowlakeNode.Generate().Int64()
	query := "INSERT INTO " + RewardReferTableName +
		" (id, user_id, win_amt, reward, reward_lv, reward_rate, data, from_unix, to_unix, send_to_wallet, create_time, update_time) " +
		" SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, 0, now(), now()" +
		" WHERE NOT EXISTS (SELECT id FROM " + RewardReferTableName + " WHERE user_id=$10 AND from_unix=$11 AND to_unix=$12) "
	l := pb.ListRewardRefer{
		UserRefers: reward.UserRefers,
	}
	data, err := conf.Marshaler.Marshal(&l)
	if err != nil {
		logger.Error("Marshalerd data error %s", err.Error())
		return 0, err
	}
	result, err := db.ExecContext(ctx, query,
		dbId, reward.GetUserId(), reward.GetWinAmt(),
		reward.GetEstReward(), reward.GetEstRewardLv(), reward.GetEstRateReward(),
		data, reward.GetFromUnix(), reward.GetToUnix(),
		reward.GetUserId(), reward.GetFromUnix(), reward.GetToUnix())
	if err != nil {
		logger.Error("Error when update reward refer user, user : %s, error %s",
			reward.UserId, err.Error())
		return 0, status.Error(codes.Internal, "Error  update reward refer user.")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Do update reward refer user, user : %s",
			reward.UserId)
		// return 0, status.Error(codes.Internal, "Error update reward refer user.")
		return UpdateRewardRefer(ctx, logger, db, reward)
	}
	return dbId, nil
}

func UpdateRewardRefer(ctx context.Context, logger runtime.Logger, db *sql.DB, reward *pb.RewardRefer) (int64, error) {
	dbId := conf.SnowlakeNode.Generate().Int64()
	query := "UPDATE  " + RewardReferTableName +
		" SET win_amt=$1, reward=$2, reward_lv=$3, reward_rate=$4, data=$5, update_time=now() " +
		" WHERE user_id=$6 AND from_unix=$7 AND to_unix=$8 AND send_to_wallet=$9"
	l := pb.ListRewardRefer{
		UserRefers: reward.UserRefers,
	}
	data, err := conf.Marshaler.Marshal(&l)
	if err != nil {
		logger.Error("Marshalerd data error %s", err.Error())
		return 0, err
	}
	result, err := db.ExecContext(ctx, query,
		reward.GetWinAmt(), reward.GetEstReward(), reward.EstRewardLv, reward.GetEstRateReward(), data,
		reward.GetUserId(), reward.GetFromUnix(), reward.GetToUnix(),
		0)
	if err != nil {
		logger.Error("Error when update reward refer user, user : %s, error %s",
			reward.UserId, err.Error())
		return 0, status.Error(codes.Internal, "Error update reward refer user.")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not update reward refer user, user : %s",
			reward.UserId)
		return 0, status.Error(codes.Internal, "Error update reward refer user.")
	}
	return dbId, nil
}

func GetRewardRefer(ctx context.Context, logger runtime.Logger, db *sql.DB, req *pb.HistoryRewardRequest) (*pb.RewardRefer, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing user id")
	}
	if req.GetFrom() <= 0 || req.GetTo() <= 0 || req.GetFrom() >= req.GetTo() {
		return nil, status.Error(codes.InvalidArgument, "Invalid time")
	}
	query := "SELECT id, user_id, win_amt, reward,reward_lv, reward_rate, data, from_unix, to_unix FROM " +
		RewardReferTableName + " WHERE user_id=$1 AND from_unix >= $2 AND to_unix <= $3"
	var dbID int64
	var dbUserId, dbData string
	var dbWinAmt, dbReward, dbRewardLv int64
	var dbRewardRate float64
	var dbFrom, dbTo int64
	err := db.QueryRowContext(ctx, query,
		req.GetUserId(), req.From,
		req.To).Scan(&dbID, &dbUserId, &dbWinAmt, &dbReward, &dbRewardLv, &dbRewardRate, &dbData, &dbFrom, &dbTo)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return &pb.RewardRefer{}, nil
		}
		logger.Error("Query history reward refer user %s error %s", req.GetUserId(), err.Error())
		return nil, status.Error(codes.Internal, "Query history reward refer error")
	}

	r := &pb.RewardRefer{
		UserId:        dbUserId,
		WinAmt:        dbWinAmt,
		EstReward:     dbReward,
		FromUnix:      dbFrom,
		ToUnix:        dbTo,
		EstRateReward: float32(dbRewardRate),
	}
	l := pb.ListRewardRefer{}
	if conf.Unmarshaler.Unmarshal([]byte(dbData), &l) == nil {
		r.UserRefers = l.GetUserRefers()
	}
	return r, nil
}

func GetHistoryRewardRefer(ctx context.Context, logger runtime.Logger, db *sql.DB, req *pb.HistoryRewardRequest) ([]*pb.RewardRefer, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing user id")
	}
	if req.GetFrom() <= 0 || req.GetTo() <= 0 || req.GetFrom() >= req.GetTo() {
		return nil, status.Error(codes.InvalidArgument, "Invalid time")
	}
	query := "SELECT id, user_id, win_amt, reward,reward_lv, reward_rate, data, from_unix, to_unix FROM " +
		RewardReferTableName + " WHERE user_id=$1 AND send_to_wallet=$2 AND from_unix >= $3 AND to_unix <= $4"
	rows, err := db.QueryContext(ctx, query,
		req.GetUserId(), 1, req.From,
		req.To)
	if err != nil {
		logger.Error("Query history reward refer user %s error %s", req.GetUserId(), err.Error())
		return nil, status.Error(codes.Internal, "Query history reward refer error")
	}
	ml := make([]*pb.RewardRefer, 0)
	var dbID int64
	var dbUserId, dbData string
	var dbWinAmt, dbReward, dbRewardLv int64
	var dbRewardRate float64
	var dbFrom, dbTo int64
	for rows.Next() {
		if rows.Scan(&dbID, &dbUserId, &dbWinAmt, &dbReward, &dbRewardLv, &dbRewardRate, &dbData, &dbFrom, &dbTo) == nil {
			r := &pb.RewardRefer{
				UserId:        dbUserId,
				WinAmt:        dbWinAmt,
				EstReward:     dbReward,
				EstRewardLv:   dbRewardLv,
				EstRateReward: float32(dbRewardRate),
				FromUnix:      dbFrom,
				ToUnix:        dbTo,
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
