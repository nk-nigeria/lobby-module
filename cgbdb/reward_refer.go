package cgbdb

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgtype"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	"github.com/nakamaFramework/cgb-lobby-module/entity"
	pb "github.com/nakamaFramework/cgp-common/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CREATE TABLE
//   public.reward_refer (
//     id bigint NOT NULL,
//     user_id character varying(128) NOT NULL,
//     win_amt bigint NOT NULL,
//     reward bigint NOT NULL,
//     reward_lv integer NOT NULL,
//     reward_rate double precision NOT NULL DEFAULT 0,
//     data character varying NULL,
//     time_send_to_wallet timestamp
//     with
//       time zone NULL,
//       from_unix bigint NULL,
//       to_unix bigint NULL,
//       create_time timestamp
//     with
//       time zone NOT NULL DEFAULT now(),
//       update_time timestamp
//     with
//       time zone NOT NULL DEFAULT now()
//   );

// ALTER TABLE
//   public.reward_refer
// ADD
//   CONSTRAINT reward_refer_pkey PRIMARY KEY (id)

const RewardReferTableName = "reward_refer"

func AddOrUpdateIfExistRewardRefer(ctx context.Context, logger runtime.Logger, db *sql.DB, reward *pb.RewardRefer) (int64, error) {
	dbId := conf.SnowlakeNode.Generate().Int64()
	query := "INSERT INTO " + RewardReferTableName +
		" (id, user_id, win_amt, reward, reward_lv, reward_rate, data, from_unix, to_unix, create_time, update_time) " +
		" SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now()" +
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

		// return 0, status.Error(codes.Internal, "Error update reward refer user.")
		return UpdateRewardRefer(ctx, logger, db, reward)
	}
	return dbId, nil
}

func UpdateRewardRefer(ctx context.Context, logger runtime.Logger, db *sql.DB, reward *pb.RewardRefer) (int64, error) {
	dbId := conf.SnowlakeNode.Generate().Int64()
	query := "UPDATE  " + RewardReferTableName +
		" SET win_amt=$1, reward=$2, reward_lv=$3, reward_rate=$4, data=$5, update_time=now() " +
		" WHERE user_id=$6 AND from_unix=$7 AND to_unix=$8 AND time_send_to_wallet is null"
	l := pb.ListRewardRefer{
		UserRefers: reward.UserRefers,
	}
	data, err := conf.Marshaler.Marshal(&l)
	if err != nil {
		logger.Error("Marshalerd data error %s", err.Error())
		return 0, err
	}
	logger.Error("Do update reward refer user, user : %s, SET win_amt= %d, reward=%d, reward_lv=%d, reward_rate=%f, data=%s",
		reward.UserId, reward.GetWinAmt(), reward.GetEstReward(),
		reward.GetEstRewardLv(), reward.GetEstRateReward(), string(data))
	result, err := db.ExecContext(ctx, query,
		reward.GetWinAmt(), reward.GetEstReward(), reward.EstRewardLv, reward.GetEstRateReward(), data,
		reward.GetUserId(), reward.GetFromUnix(), reward.GetToUnix())
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

func GetRewardRefer(ctx context.Context, logger runtime.Logger, db *sql.DB, req *entity.FeeGameListCursor) (*pb.RewardRefer, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing user id")
	}
	if req.From <= 0 || req.To <= 0 || req.From >= req.To {
		return nil, status.Error(codes.InvalidArgument, "Invalid time")
	}
	query := "SELECT id, user_id, win_amt, reward,reward_lv, reward_rate, data, from_unix, to_unix, create_time, update_time FROM " +
		RewardReferTableName + " WHERE user_id=$1 AND from_unix >= $2 AND to_unix <= $3"
	var dbID int64
	var dbUserId, dbData string
	var dbWinAmt, dbReward, dbRewardLv int64
	var dbRewardRate float64
	var dbFrom, dbTo int64
	var dbCreateTime, dbUpdateTime pgtype.Timestamptz
	err := db.QueryRowContext(ctx, query,
		req.UserId, req.From,
		req.To).Scan(&dbID, &dbUserId, &dbWinAmt, &dbReward, &dbRewardLv, &dbRewardRate, &dbData, &dbFrom, &dbTo, &dbCreateTime, &dbUpdateTime)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return &pb.RewardRefer{}, nil
		}
		logger.Error("Query history reward refer user %s error %s", req.UserId, err.Error())
		return nil, status.Error(codes.Internal, "Query history reward refer error")
	}

	r := &pb.RewardRefer{
		Id:             dbID,
		UserId:         dbUserId,
		WinAmt:         dbWinAmt,
		EstReward:      dbReward,
		FromUnix:       dbFrom,
		ToUnix:         dbTo,
		EstRateReward:  float32(dbRewardRate),
		CreateTimeUnix: dbCreateTime.Time.Unix(),
		UpdateTimeUnix: dbUpdateTime.Time.Unix(),
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
		RewardReferTableName + " WHERE user_id=$1 AND time_send_to_wallet IS NOT NULL AND from_unix >= $2 AND to_unix <= $3"
	rows, err := db.QueryContext(ctx, query,
		req.GetUserId(), req.From, req.To)
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
				Id:            dbID,
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

func GetListRewardCompleteReferNotSendToWallet(ctx context.Context, logger runtime.Logger, db *sql.DB, limit int64, offset int64) ([]*pb.RewardRefer, error) {
	query := "SELECT id, user_id, win_amt, reward,reward_lv, reward_rate, data, from_unix, to_unix FROM " +
		RewardReferTableName + " WHERE time_send_to_wallet IS NULL AND to_unix <= $1 limit $2 offset $3"
	rows, err := db.QueryContext(ctx, query, time.Now().Unix(), limit, offset)
	if err != nil {
		logger.Error("Query list reward refer not send to wallet err %s", err.Error())
		return nil, status.Error(codes.Internal, "Query list reward refer not send to wallet error")
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
				Id:            dbID,
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

func GetListRewardReferNotComplete(ctx context.Context, logger runtime.Logger, db *sql.DB, limit int64, offset int64) ([]*pb.RewardRefer, error) {
	_, endLastWeek := entity.RangeLastWeek()
	query := "SELECT id, user_id, win_amt, reward,reward_lv, reward_rate, data, from_unix, to_unix FROM " +
		RewardReferTableName + " WHERE time_send_to_wallet IS NULL AND to_unix <= $1 AND update_time < $2 limit $3 offset $4"
	rows, err := db.QueryContext(ctx, query, endLastWeek.Unix(), endLastWeek, limit, offset)
	if err != nil {
		logger.Error("Query list reward refer not complete err %s", err.Error())
		return nil, status.Error(codes.Internal, "Query list reward refer not complete error")
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
				Id:            dbID,
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

func UpdateRewardReferHasSendToWallet(ctx context.Context, logger runtime.Logger, db *sql.DB, rewardReferID int64) error {
	query := "UPDATE  " + RewardReferTableName +
		" SET time_send_to_wallet=now() " +
		" WHERE id=$1 AND time_send_to_wallet is null"
	result, err := db.ExecContext(ctx, query,
		rewardReferID)
	if err != nil {
		logger.Error("Error when update reward refer have send to wallet user, id %d, error %s",
			rewardReferID, err.Error())
		return status.Error(codes.Internal, "Error update reward has send to refer user.")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not update reward has send to refer user, id", rewardReferID)
		return status.Error(codes.Internal, "Error update reward has send to refer user.")
	}
	return nil
}
