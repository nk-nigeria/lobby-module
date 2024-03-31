package api

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/ciaolink-game-platform/cgb-lobby-module/cgbdb"
	"github.com/ciaolink-game-platform/cgp-common/define"
	api "github.com/ciaolink-game-platform/cgp-common/proto"
	nkapi "github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

func CustomEventHandler(db *sql.DB) func(ctx context.Context, logger runtime.Logger, evt *nkapi.Event) {
	return func(ctx context.Context, logger runtime.Logger, evt *nkapi.Event) {
		if evt == nil {
			return
		}
		eventName := evt.GetName()
		// event end of match
		switch eventName {
		case string(define.NakEventMatchEnd):
			eventNakamaMatchEnd(ctx, logger, db, evt)
		case string(define.NakEventMatchJoin):
			eventNakamaMatchJoin(ctx, logger, db, evt)
		case string(define.NakEventMatchLeave):
			eventNakamaMatchLeave(ctx, logger, db, evt)
		default:
			return
		}
	}
}

func eventNakamaMatchEnd(ctx context.Context, logger runtime.Logger, db *sql.DB, evt *nkapi.Event) {
	userId := evt.Properties["user_id"]
	gameCode := evt.Properties["game_code"]
	tsEndStr := evt.Properties["end_match_unix"]
	tsEndUnix, _ := strconv.ParseInt(tsEndStr, 10, 64)
	matchId := ""
	mcb, _ := strconv.ParseInt(evt.Properties["mcb"], 10, 64)
	lastBet, _ := strconv.ParseInt(evt.Properties["last_bet"], 10, 64)
	cgbdb.UpdateUsersPlayingInMatch(ctx, logger, db, userId, &api.PlayingMatch{
		Code:      gameCode,
		MatchId:   matchId,
		LeaveTime: tsEndUnix,
		Mcb:       mcb,
		Bet:       lastBet,
	})
}

func eventNakamaMatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, evt *nkapi.Event) {
	userId := evt.Properties["user_id"]
	gameCode := evt.Properties["game_code"]
	tsEndStr := evt.Properties["end_match_unix"]
	tsEndUnix, _ := strconv.ParseInt(tsEndStr, 10, 64)
	matchId := evt.Properties["match_id"]
	mcb, _ := strconv.ParseInt(evt.Properties["mcb"], 10, 64)
	lastBet, _ := strconv.ParseInt(evt.Properties["last_bet"], 10, 64)
	cgbdb.UpdateUsersPlayingInMatch(ctx, logger, db, userId, &api.PlayingMatch{
		Code:      gameCode,
		MatchId:   matchId,
		LeaveTime: tsEndUnix,
		Mcb:       mcb,
		Bet:       lastBet,
	})
}

func eventNakamaMatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, evt *nkapi.Event) {
	userId := evt.Properties["user_id"]
	gameCode := evt.Properties["game_code"]
	tsEndStr := evt.Properties["end_match_unix"]
	tsEndUnix, _ := strconv.ParseInt(tsEndStr, 10, 64)
	matchId := evt.Properties["match_id"]
	mcb, _ := strconv.ParseInt(evt.Properties["mcb"], 10, 64)
	lastBet, _ := strconv.ParseInt(evt.Properties["last_bet"], 10, 64)
	cgbdb.UpdateUsersPlayingInMatch(ctx, logger, db, userId, &api.PlayingMatch{
		Code:      gameCode,
		MatchId:   matchId,
		LeaveTime: tsEndUnix,
		Mcb:       mcb,
		Bet:       lastBet,
	})
}
