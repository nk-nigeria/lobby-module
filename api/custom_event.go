package api

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"sync"

	nkapi "github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgb-lobby-module/cgbdb"
	"github.com/nakamaFramework/cgp-common/define"
	api "github.com/nakamaFramework/cgp-common/proto"
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
	userIds := evt.Properties["user_id"]
	if len(userIds) == 0 {
		return
	}
	gameCode := evt.Properties["game_code"]
	tsEndStr := evt.Properties["end_match_unix"]
	tsEndUnix, _ := strconv.ParseInt(tsEndStr, 10, 64)
	matchId := ""
	mcb, _ := strconv.ParseInt(evt.Properties["mcb"], 10, 64)
	lastBet, _ := strconv.ParseInt(evt.Properties["last_bet"], 10, 64)
	mt.Lock()
	defer mt.Unlock()
	for _, userId := range strings.Split(userIds, ",") {
		cgbdb.UpdateUsersPlayingInMatch(ctx, logger, db, userId, &api.PlayingMatch{
			Code:      gameCode,
			MatchId:   matchId,
			LeaveTime: tsEndUnix,
			Mcb:       mcb,
			Bet:       lastBet,
		})
		trackGame, exist := trackUserInGame[gameCode]
		if !exist {
			trackGame = make(playerByMcb)
		}
		playerByMcb := trackGame[int(mcb)]
		delete(playerByMcb, userId)
		trackGame[int(mcb)] = playerByMcb
		trackUserInGame[gameCode] = trackGame
	}

}

func eventNakamaMatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, evt *nkapi.Event) {
	userIds := evt.Properties["user_id"]
	if len(userIds) == 0 {
		return
	}
	gameCode := evt.Properties["game_code"]
	tsEndStr := evt.Properties["end_match_unix"]
	tsEndUnix, _ := strconv.ParseInt(tsEndStr, 10, 64)
	matchId := evt.Properties["match_id"]
	mcb, _ := strconv.ParseInt(evt.Properties["mcb"], 10, 64)
	lastBet, _ := strconv.ParseInt(evt.Properties["last_bet"], 10, 64)
	mt.Lock()
	defer mt.Unlock()
	for _, userId := range strings.Split(userIds, ",") {
		cgbdb.UpdateUsersPlayingInMatch(ctx, logger, db, userId, &api.PlayingMatch{
			Code:      gameCode,
			MatchId:   matchId,
			LeaveTime: tsEndUnix,
			Mcb:       mcb,
			Bet:       lastBet,
		})
		trackGame, exist := trackUserInGame[gameCode]
		if !exist {
			trackGame = make(playerByMcb)
		}
		playerByMcb := trackGame[int(mcb)]
		if playerByMcb == nil {
			playerByMcb = make(players)
		}
		playerByMcb[userId] = struct{}{}
		trackGame[int(mcb)] = playerByMcb
		trackUserInGame[gameCode] = trackGame
	}
}

func eventNakamaMatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, evt *nkapi.Event) {
	userIds := evt.Properties["user_id"]
	if len(userIds) == 0 {
		return
	}

	gameCode := evt.Properties["game_code"]
	tsEndStr := evt.Properties["end_match_unix"]
	tsEndUnix, _ := strconv.ParseInt(tsEndStr, 10, 64)
	matchId := evt.Properties["match_id"]
	mcb, _ := strconv.ParseInt(evt.Properties["mcb"], 10, 64)
	lastBet, _ := strconv.ParseInt(evt.Properties["last_bet"], 10, 64)
	mt.Lock()
	defer mt.Unlock()
	for _, userId := range strings.Split(userIds, ",") {
		cgbdb.UpdateUsersPlayingInMatch(ctx, logger, db, userId, &api.PlayingMatch{
			Code:      gameCode,
			MatchId:   matchId,
			LeaveTime: tsEndUnix,
			Mcb:       mcb,
			Bet:       lastBet,
		})

		trackGame, exist := trackUserInGame[gameCode]
		if exist {
			playerByMcb := trackGame[int(mcb)]
			delete(playerByMcb, userId)
			trackGame[int(mcb)] = playerByMcb
			trackUserInGame[gameCode] = trackGame
		}
	}

}
