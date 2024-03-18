// Copyright 2020 The Nakama Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"

	"github.com/ciaolink-game-platform/cgp-common/define"

	"github.com/ciaolink-game-platform/cgb-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgb-lobby-module/cgbdb"
	"github.com/ciaolink-game-platform/cgb-lobby-module/conf"
	"github.com/ciaolink-game-platform/cgb-lobby-module/entity"
	pb "github.com/ciaolink-game-platform/cgp-common/proto"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

const kDefaultMaxSize = 3

type MatchLabel struct {
	Open         int32  `json:"open"`
	Mcb          int32  `json:"mcb"`
	Bet          int64  `json:"bet"`
	LastBet      int64  `json:"last_bet"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Password     string `json:"password"`
	MaxSize      int32  `json:"max_size"`
	MockCodeCard int32  `json:"mock_code_card"`
}

var GetTableId func() string

func init() {
	// GetTableId = func() int64 {
	// 	var counter atomic.Int64
	// 	counter.Store(0)
	// 	return func() int64 {
	// 		newVal := counter.Add(1)
	// 		return newVal
	// 	}
	// }
	GetTableId = func() func() string {
		// var counter atomic.Int64 // feature only available on go 1.19
		var counter int64 = 0
		var mt sync.Mutex
		return func() string {
			mt.Lock()
			counter += 1
			newVal := counter
			mt.Unlock()
			return fmt.Sprintf("%05d", newVal)
		}
	}()
}

func RpcFindMatch(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		emptyResp := "[]"
		logger.Info("rpc find match: %v", payload)
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		request := &pb.RpcFindMatchRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			return "", presenter.ErrUnmarshal
		}
		if err := checkEnoughChipForBet(ctx, logger, db, nk, userID, request.GameCode, int64(request.MarkUnit), false); err != nil {
			return "", err
		}

		maxSize := kDefaultMaxSize

		var query string
		if !request.WithNonOpen {
			query = fmt.Sprintf("+label.name:%s +label.markUnit:%d", request.GameCode, request.MarkUnit)
		} else {
			query = fmt.Sprintf("+label.open:>0 +label.name:%s +label.markUnit:%d", request.GameCode, request.MarkUnit)
		}

		// request.MockCodeCard = 0
		if request.MockCodeCard > 0 {
			query += fmt.Sprintf(" +label.mock_code_card:%d", request.MockCodeCard)
		}

		logger.Info("match query %v", query)

		resMatches := &pb.RpcFindMatchResponse{}
		matches, err := nk.MatchList(ctx, 10, true, "", nil, &maxSize, query)
		if err != nil {
			logger.Error("error listing matches: %v", err)
			return "", presenter.ErrInternalError
		}

		logger.Debug("find match result %v", matches)

		for _, match := range matches {
			// var label MatchLabel
			matchInfo := &pb.Match{}
			err = conf.Unmarshaler.Unmarshal([]byte(match.Label.GetValue()), matchInfo)
			if err != nil {
				logger.Error("unmarshal label error %v", err)
				continue
			}

			// logger.Debug("find match size: %v", match.Size)
			if matchInfo.Size >= matchInfo.MaxSize {
				continue
			}
			resMatches.Matches = append(resMatches.Matches, matchInfo)
		}
		if len(resMatches.Matches) <= 0 && request.Create {
			resMatches.Matches, err = createMatch(ctx, logger, db, nk, &pb.RpcCreateMatchRequest{
				GameCode: request.GameCode,
				MarkUnit: request.MarkUnit,
				MaxSize:  int64(maxSize),
				Password: request.GetPassword(),
			})
			if err != nil {
				logger.WithField("err", err).Error("error creating match")
				return "", presenter.ErrInternalError
			}
			response, err := conf.MarshalerDefault.Marshal(resMatches)
			if err != nil {
				logger.Error("error marshaling response payload: %v", err.Error())
				return "", presenter.ErrMarshal
			}
			return string(response), nil
		}
		//  not found match,
		if len(resMatches.Matches) <= 0 {
			logger.Error("Not found match for user %s", userID)
			return emptyResp, nil
		}
		for _, match := range resMatches.Matches {
			match.NumBot = 0
			match.MockCodeCard = 0
			match.Open = len(match.Password) > 0
		}

		response, err := marshaler.Marshal(resMatches)
		if err != nil {
			logger.Error("error marshaling response payload: %v", err.Error())
			return "", presenter.ErrMarshal
		}

		return string(response), nil
	}
}

func RpcQuickMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	defer Recovery(logger)
	logger.Info("rpc quick match: %v", payload)
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		return "", presenter.ErrNoUserIdFound
	}
	unmarshaler := conf.Unmarshaler
	marshaler := conf.MarshalerDefault

	request := &pb.RpcCreateMatchRequest{}
	if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
		logger.Error("error unmarhal input : %v", err)
		return "", presenter.ErrUnmarshal
	}
	if len(request.GameCode) == 0 {
		// return "", presenter.ErrInvalidInput
		return quickMatchAtLobby(ctx, logger, db, nk, request)
	}
	if err := checkEnoughChipForBet(ctx, logger, db, nk, userID, request.GameCode, int64(request.MarkUnit), true); err != nil {
		logger.WithField("user", userID).WithField("min chip", request.MarkUnit).WithField("err", err).Error("not enough chip for bet")
		return "", presenter.ErrNotEnoughChip
	}
	maxSize := kDefaultMaxSize
	query := fmt.Sprintf("+label.code:%s +label.open:1", request.GameCode)

	resMatches := &pb.RpcFindMatchResponse{}
	var matches []*api.Match
	var err error
	if define.IsAllowJoinInGameOnProgress(request.GameCode) {
		matches, err = nk.MatchList(ctx, 100, true, "", nil, &maxSize, query)
		if err != nil {
			logger.Error("error listing matches: %v", err)
			return "", presenter.ErrInternalError
		}

		logger.Debug("MatchList result %v", matches)
	}
	// matchInfo := &pb.Match{
	// 	Size:     1,
	// 	MaxSize:  int32(maxSize),
	// 	Name:     request.Name,
	// 	Open:     len(request.Password) > 0,
	// 	LastBet:  request.LastBet,
	// 	TableId:  GetTableId(),
	// 	Password: request.Password,
	// 	NumBot:   1,
	// 	// UserCreated: ,
	// }
	if len(matches) == 0 {
		resMatches.Matches, err = createMatch(ctx, logger, db, nk, request)

	}
	// There are one or more ongoing matches the user could join.
	for _, match := range matches {
		// var label MatchLabel
		mInfo := &pb.Match{}
		err = unmarshaler.Unmarshal([]byte(match.Label.GetValue()), mInfo)
		if err != nil {
			logger.Error("unmarshal label error %v", err)
			continue
		}

		logger.Debug("find match %v", match.Size)
		resMatches.Matches = append(resMatches.Matches, mInfo)
	}

	sort.Slice(resMatches.Matches, func(i, j int) bool {
		r := resMatches.Matches[i].MarkUnit < resMatches.Matches[j].MarkUnit
		return r
	})
	response, err := marshaler.Marshal(resMatches)
	if err != nil {
		logger.Error("error marshaling response payload: %v", err.Error())
		return "", presenter.ErrMarshal
	}

	return string(response), nil
}

func RpcCreateMatch(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		logger.Info("rpc create match: %v", payload)

		_, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		request := &pb.RpcCreateMatchRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			logger.Error("unmarshal create match error %v", err)
			return "", presenter.ErrUnmarshal
		}
		matchs, err := createMatch(ctx, logger, db, nk, request)
		if err != nil {
			logger.WithField("err", err).Error("error creating match")
			return "", presenter.ErrInternalError
		}
		response, err := conf.MarshalerDefault.Marshal(&pb.RpcCreateMatchResponse{MatchId: matchs[0].MatchId})
		if err != nil {
			logger.Error("error marshaling response payload: %v", err.Error())
			return "", presenter.ErrMarshal
		}
		return string(response), nil
	}
}

func createMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, request *pb.RpcCreateMatchRequest) ([]*pb.Match, error) {
	defer Recovery(logger)
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		return nil, presenter.ErrNoUserIdFound
	}
	// if request.MarkUnit <0 {
	// 	return nil, presenter.ErrNoInputAllowed
	// }
	account, _, err := cgbdb.GetProfileUser(ctx, db, userID, nil)
	if err != nil {
		logger.WithField("err", err).Error("get account failed")
		return nil, err
	}
	matchInfo := &pb.Match{
		Size:     1,
		MaxSize:  int32(request.MaxSize),
		Name:     request.Name,
		Open:     len(request.Password) > 0,
		LastBet:  request.LastBet,
		TableId:  GetTableId(),
		Password: request.Password,
		NumBot:   1,
		MarkUnit: request.MarkUnit,
		UserCreated: &pb.Profile{
			UserId:      account.UserId,
			UserSid:     account.UserSid,
			UserName:    account.UserName,
			DisplayName: account.DisplayName,
		},
	}
	if len(matchInfo.Name) == 0 {
		matchInfo.Name = request.GameCode
	}
	if matchInfo.MaxSize <= 0 {
		matchInfo.MaxSize = kDefaultMaxSize
	}
	if matchInfo.NumBot <= 0 {
		matchInfo.NumBot = 1
	}
	bets, err := LoadBets(ctx, logger, db, nk, request.GameCode)
	if err != nil {
		return nil, presenter.ErrInternalError
	}
	if len(bets) > 0 {
		sort.Slice(bets, func(i, j int) bool {
			return bets[i].MarkUnit < bets[j].MarkUnit
		})
		if err := checkEnoughChipForBet(ctx, logger, db, nk, userID, request.GameCode, int64(bets[0].MarkUnit), true); err != nil {
			logger.WithField("user", userID).WithField("min chip", request.MarkUnit).WithField("err", err).Error("not enough chip for bet")
			return nil, err
		}
	} else {
		game, ok := mapGameByCode.Load(request.GameCode)
		gameId := game.(entity.Game).ID
		if !ok {
			logger.WithField("gamecode", request.GameCode).Error("not found game")
			return nil, presenter.ErrMatchNotFound
		}
		bet := entity.Bet{
			MarkUnit: int(request.GetMarkUnit()),
			Enable:   true,
			Id:       0,
			GameId:   int(gameId),
		}
		bets = append(bets, bet)
	}
	// No available matches found, create a new one.
	arg := make(map[string]any)
	matchInfo.TableId = GetTableId()
	if len(bets) == 0 && matchInfo.MarkUnit <= 0 {
		return nil, presenter.ErrNoInputAllowed
	}
	// check bet in list config bet
	if len(bets) > 0 {
		if matchInfo.MarkUnit == 0 {
			matchInfo.MarkUnit = int32(bets[0].MarkUnit)
		}
		validMarkUnit := false
		for _, bet := range bets {
			if bet.MarkUnit == int(matchInfo.MarkUnit) {
				validMarkUnit = true
			}
		}
		if !validMarkUnit {
			return nil, presenter.ErrNoInputAllowed
		}
	}
	data, _ := conf.MarshalerDefault.Marshal(matchInfo)
	arg["data"] = string(data)
	matchID, err := nk.MatchCreate(ctx, request.GameCode, arg)
	if err != nil {
		logger.WithField("data", data).Error("error creating match: %v", err)
		return nil, presenter.ErrInternalError
	}
	matchInfo.MatchId = matchID
	matchInfo.NumBot = 0
	matchInfo.MockCodeCard = 0
	matchInfo.Open = len(matchInfo.Password) > 0
	resMatches := &pb.RpcFindMatchResponse{}
	resMatches.Matches = append(resMatches.Matches, matchInfo)
	return resMatches.Matches, nil
}
func checkEnoughChipForBet(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, userID string, gameCode string, betWantCheck int64, quickJoin bool) error {
	bets, err := LoadBets(ctx, logger, db, nk, gameCode)
	if err != nil {
		return presenter.ErrInternalError
	}
	if len(bets) == 0 {
		return nil
	}
	var bet entity.Bet
	for _, v := range bets {
		if betWantCheck == 0 {
			bet = v
			break
		}
		if v.MarkUnit == int(betWantCheck) {
			bet = v
			break
		}
	}
	if bet.MarkUnit <= 0 {
		return presenter.ErrBetNotFound
	}
	minChipRequire := bet.AGJoin
	if quickJoin {
		minChipRequire = bet.AGPlaynow
	}
	wallet, err := entity.ReadWalletUser(ctx, nk, logger, userID)
	if err != nil {
		logger.Error("read wallet user %s error %s",
			userID, err.Error())
		return presenter.ErrInternalError
	}
	if wallet.Chips <= 0 || wallet.Chips < int64(minChipRequire) {
		logger.Error("User %s not enough chip [%d] to join game bet [%d]",
			userID, wallet.Chips, bet)
		return presenter.ErrNotEnoughChip
	}
	return nil
}

func quickMatchAtLobby(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, req *pb.RpcCreateMatchRequest) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
	}
	gameCode := "gaple" // default game
	profile, _, err := cgbdb.GetProfileUser(ctx, db, userID, nil)
	if err != nil {
		logger.WithField("user id", userID).WithField("err", err).Error("get profile failed")
		return "", err
	}
	if len(profile.PlayingMatch.Code) != 0 {
		gameCode = profile.PlayingMatch.Code
		req.MarkUnit = int32(profile.PlayingMatch.Mcb)
		req.LastBet = profile.PlayingMatch.Bet
	} else {
		req.MarkUnit = 0
	}
	req.GameCode = gameCode
	return RpcQuickMatch(ctx, logger, db, nk, req.String())
}
