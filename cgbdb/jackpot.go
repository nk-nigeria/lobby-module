package cgbdb

import (
	"context"
	"database/sql"
	"strings"

	pb "github.com/ciaolink-game-platform/cgp-common/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CREATE TABLE
//   public.jackpot (
//     id bigint NOT NULL,
// 	   game character varying(128) NOT NULL,
//     UNIQUE(game),
//     chips bigint NOT NULL DEFAULT 0,
//     create_time timestamp
//     with
//       time zone NOT NULL DEFAULT now(),
//       update_time timestamp
//     with
//       time zone NOT NULL DEFAULT now()
//   );

// ALTER TABLE
//   public.jackpot
// ADD
//   CONSTRAINT jackpot_pkey PRIMARY KEY (id)

// CREATE TABLE
//   public.jackpot_history (
//     id bigint NOT NULL,
// 	   game character varying(128) NOT NULL,
//     chips bigint NOT NULL DEFAULT 0,
//     metadata string
//     create_time timestamp
//     with
//       time zone NOT NULL DEFAULT now(),
//       update_time timestamp
//     with
//       time zone NOT NULL DEFAULT now()
//   );

// ALTER TABLE
//   public.jackpot_history
// ADD
//   CONSTRAINT jackpot_pkey PRIMARY KEY (id)

const JackpotTableName = "jackpot"

func GetJackpot(ctx context.Context, logger runtime.Logger, db *sql.DB, game string) (*pb.Jackpot, error) {
	query := "SELECT id, game, chips, create_time FROM " + JackpotTableName +
		" WHERE game=$1"
	var dbId, dbChips int64
	var dbGame string
	var dbCreateTime pgtype.Timestamptz
	err := db.QueryRowContext(ctx, query, game).
		Scan(&dbId, &dbGame, &dbChips, &dbCreateTime)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return &pb.Jackpot{}, nil
		}
		logger.WithField("game", game).WithField("err", err.Error()).Error("Query jackpot error")
		return nil, status.Error(codes.Internal, "Query jackpot error")
	}
	jackpot := &pb.Jackpot{
		Id:             dbId,
		GameCode:       dbGame,
		Chips:          dbChips,
		CreateTimeUnix: dbCreateTime.Time.Unix(),
	}
	return jackpot, nil
}
