package api

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/ciaolink-game-platform/cgb-lobby-module/cgbdb"
	"github.com/ciaolink-game-platform/cgp-common/define"
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
	cgbdb.UpdateUsersPlayingInMatch(ctx, logger, db, userId, matchId, gameCode, tsEndUnix)
}

func eventNakamaMatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, evt *nkapi.Event) {
	userId := evt.Properties["user_id"]
	gameCode := evt.Properties["game_code"]
	tsEndStr := evt.Properties["end_match_unix"]
	tsEndUnix, _ := strconv.ParseInt(tsEndStr, 10, 64)
	matchId := evt.Properties["match_id"]
	cgbdb.UpdateUsersPlayingInMatch(ctx, logger, db, userId, matchId, gameCode, tsEndUnix)
}

func eventNakamaMatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, evt *nkapi.Event) {
	userId := evt.Properties["user_id"]
	gameCode := evt.Properties["game_code"]
	tsEndStr := evt.Properties["end_match_unix"]
	tsEndUnix, _ := strconv.ParseInt(tsEndStr, 10, 64)
	matchId := ""
	cgbdb.UpdateUsersPlayingInMatch(ctx, logger, db, userId, matchId, gameCode, tsEndUnix)
}
