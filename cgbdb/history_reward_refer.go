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
//   public.history_reward_refer (
//     id bigint NOT NULL,
//     user_id character varying(128) NOT NULL,
//     win_amt bigint NOT NULL,\
//     reward bigint NOT NULL,\
//     data VARCHAR,\
// ALTER TABLE
//   public.history_reward_refer
// ADD
//   CONSTRAINT rhistory_reward_refer_pkey PRIMARY KEY (id)
const HistoryRewardReferTableName = "history_reward_refer"

func AddNewHistoryRewardRefer(ctx context.Context, logger runtime.Logger, db *sql.DB, reward *pb.RewardRefer) (int64, error) {
	dbId := conf.SnowlakeNode.Generate().Int64()
	query := "INSERT INTO " + HistoryRewardReferTableName +
		" (id, user_id, win_amt, reward, data, create_time, update_time) " +
		"VALUES ($1, $2, $3, $4, $5, now(), now())"
	l := pb.ListRewardRefer{
		UserRefers: reward.UserRefers,
	}
	data, err := conf.Marshaler.Marshal(&l)
	if err != nil {
		logger.Error("Marshalerd data error %s", err.Error())
		return 0, err
	}
	result, err := db.ExecContext(ctx, query,
		dbId, reward.UserId, reward.WinAmt, reward.EstReward, data)
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
