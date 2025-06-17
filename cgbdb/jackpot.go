package cgbdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgtype"
	pb "github.com/nk-nigeria/cgp-common/proto"
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

func GetJackpotsByGame(ctx context.Context, logger runtime.Logger, db *sql.DB, games ...string) ([]*pb.Jackpot, error) {
	query := "SELECT id, game, chips, create_time FROM " + JackpotTableName +
		fmt.Sprintf(` WHERE game IN ('%s')`, strings.Join(games, "','"))

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		logger.WithField("game", games).WithField("err", err.Error()).Error("Query jackpot error")
		return nil, status.Error(codes.Internal, "Query jackpot error")
	}
	defer rows.Close()

	var jackpots []*pb.Jackpot
	for rows.Next() {
		var dbId, dbChips int64
		var dbGame string
		var dbCreateTime pgtype.Timestamptz
		if err := rows.Scan(&dbId, &dbGame, &dbChips, &dbCreateTime); err != nil {
			logger.WithField("game", games).WithField("err", err.Error()).Error("Scan jackpot error")
			return nil, status.Error(codes.Internal, "Scan jackpot error")
		}
		jackpot := &pb.Jackpot{
			Id:             dbId,
			GameCode:       dbGame,
			Chips:          dbChips,
			CreateTimeUnix: dbCreateTime.Time.Unix(),
		}
		jackpots = append(jackpots, jackpot)
	}

	if err := rows.Err(); err != nil {
		logger.WithField("game", games).WithField("err", err.Error()).Error("Iterating rows error")
		return nil, status.Error(codes.Internal, "Iterating rows error")
	}

	return jackpots, nil
}
