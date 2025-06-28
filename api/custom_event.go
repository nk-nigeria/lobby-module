package api

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"sync"

	nkapi "github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nk-nigeria/cgp-common/define"
	api "github.com/nk-nigeria/cgp-common/proto"
	"github.com/nk-nigeria/lobby-module/cgbdb"
)

type players map[string]struct{}

type playerByMcb map[int]players // [mcb]num_player

var trackUserInGame map[string]playerByMcb = make(map[string]playerByMcb) // [gamecode]
var mt sync.Mutex

func (c playerByMcb) TotalPlayer(mcb int) int {
	player := c[mcb]
	return len(player)
}

func CustomEventHandler(db *sql.DB) func(ctx context.Context, logger runtime.Logger, evt *nkapi.Event) {
	return func(ctx context.Context, logger runtime.Logger, evt *nkapi.Event) {
		if evt == nil {
			return
		}
		eventName := evt.GetName()
		logger.Debug("CustomEventHandler", "event_name", eventName, "event_id", "event_properties", evt.GetProperties())
		// event end of match
		switch eventName {
		case string(define.NakEventMatchJoin):
			updateUserMatch(ctx, logger, db, evt, false)
		case string(define.NakEventMatchLeave):
			updateUserMatch(ctx, logger, db, evt, true)
		case string(define.NakEventMatchEnd):
			updateUserMatch(ctx, logger, db, evt, true)
		default:
			return
		}
	}
}

func updateUserMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, evt *nkapi.Event, isLeave bool) {
	userIds := strings.Split(evt.Properties["user_id"], ",")
	if len(userIds) == 0 {
		return
	}

	gameCode := evt.Properties["game_code"]
	matchId := evt.Properties["match_id"]
	tsEndUnix, _ := strconv.ParseInt(evt.Properties["end_match_unix"], 10, 64)
	mcb, _ := strconv.Atoi(evt.Properties["mcb"])
	lastBet, _ := strconv.ParseInt(evt.Properties["last_bet"], 10, 64)

	mt.Lock()
	defer mt.Unlock()

	trackGame, exist := trackUserInGame[gameCode]
	if !exist {
		trackGame = make(playerByMcb)
	}

	playerSet := trackGame[mcb]
	if playerSet == nil {
		playerSet = make(players)
	}

	for _, userId := range userIds {
		cgbdb.UpdateUsersPlayingInMatch(ctx, logger, db, userId, &api.PlayingMatch{
			Code:      gameCode,
			MatchId:   matchId,
			LeaveTime: tsEndUnix,
			Mcb:       int64(mcb),
			Bet:       lastBet,
		})

		if isLeave {
			delete(playerSet, userId)
		} else {
			playerSet[userId] = struct{}{}
		}
	}
	trackGame[mcb] = playerSet
	trackUserInGame[gameCode] = trackGame
}
