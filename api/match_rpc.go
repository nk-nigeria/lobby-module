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
	"encoding/json"
	"fmt"
	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	"sort"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
)

type MatchLabel struct {
	Open     int32  `json:"open"`
	Bet      int32  `json:"bet"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Password string `json:"password"`
	MaxSize  int32  `json:"max_size"`
}

func RpcFindMatch(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		logger.Info("rpc find match: %v", payload)
		_, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		request := &pb.RpcFindMatchRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			return "", presenter.ErrUnmarshal
		}

		maxSize := 3
		query := fmt.Sprintf("+label.code:%s +label.bet:%d", request.GameCode, request.Bet)

		resMatches := &pb.RpcFindMatchResponse{}
		matches, err := nk.MatchList(ctx, 10, true, "", nil, &maxSize, query)
		if err != nil {
			logger.Error("error listing matches: %v", err)
			return "", presenter.ErrInternalError
		}

		logger.Debug("find match result %v", matches)
		if len(matches) > 0 {
			// There are one or more ongoing matches the user could join.
			for _, match := range matches {
				var label MatchLabel
				err = json.Unmarshal([]byte(match.Label.GetValue()), &label)
				if err != nil {
					logger.Error("unmarshal label error %v", err)
					continue
				}

				logger.Debug("find match %v", match.Size)
				resMatches.Matches = append(resMatches.Matches, &pb.Match{
					MatchId: match.MatchId,
					Size:    match.Size,
					MaxSize: label.MaxSize, // Get from label
					Name:    label.Name,
					Bet: &pb.Bet{
						MarkUnit: label.Bet,
						Enable:   true,
					},
				})
			}
		}

		response, err := marshaler.Marshal(resMatches)
		if err != nil {
			logger.Error("error marshaling response payload: %v", err.Error())
			return "", presenter.ErrMarshal
		}

		return string(response), nil
	}
}

//RpcQuickMatch
//Case 1: find exists match with matching bet
//Case 2: not found any match, create new matching one witch matching bet
//Case 3: not found any match, not found any matching bet
func RpcQuickMatch(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		logger.Info("rpc quick match: %v", payload)
		uid, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		request := &pb.RpcCreateMatchRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			return "", presenter.ErrUnmarshal
		}

		// load bet config
		bets, err := LoadBets(request.GameCode, ctx, logger, nk)
		if err != nil {
			return "", presenter.ErrInternalError
		}

		if len(bets.Bets) == 0 {
			return "", nil
		}

		sort.Slice(bets.Bets, func(i, j int) bool {
			return bets.Bets[i].MarkUnit < bets.Bets[j].MarkUnit
		})

		// load user wallet
		wallets, err := entity.ReadWalletUsers(ctx, nk, logger, uid)
		if err != nil {
			logger.Warn("error wallet not found ", uid)
			return "", presenter.ErrInternalError
		}

		wallet := wallets[0]

		maxSize := 3
		query := fmt.Sprintf("+label.code:%s", request.GameCode)

		resMatches := &pb.RpcFindMatchResponse{}
		matches, err := nk.MatchList(ctx, 0, true, "", nil, &maxSize, query)
		if err != nil {
			logger.Error("error listing matches: %v", err)
			return "", presenter.ErrInternalError
		}

		var match pb.Match
		var foundBet = false
		var matchBet int32
		if len(matches) == 0 {
			logger.Info("don't found any match, try to create one")
			for _, bet := range bets.Bets {
				if wallet.Chips >= int64(bet.AGPlaynow) {
					foundBet = true
					matchBet = bet.MarkUnit
					break
				}
			}

			if foundBet {
				// No available matches found, create a new one.
				match.MatchId, err = nk.MatchCreate(ctx, request.GameCode, map[string]interface{}{
					"bet":       matchBet,
					"game_code": request.GameCode,
					"name":      request.Name,
					"password":  request.Password,
				})

				if err != nil {
					logger.Error("error creating match: %v", err)
					return "", presenter.ErrInternalError
				}

				match.Bet = &pb.Bet{
					MarkUnit: matchBet,
				}

				match.Size = 1
			}
		} else {
			logger.Info("find matching bet match")
			sort.Slice(matches, func(i, j int) bool {
				var l1, l2 MatchLabel
				if err := json.Unmarshal([]byte(matches[i].Label.GetValue()), l1); err != nil {

				}

				if err := json.Unmarshal([]byte(matches[j].Label.GetValue()), l2); err != nil {

				}

				return l1.Bet < l2.Bet
			})

			for _, pmatch := range matches {
				var label MatchLabel
				if err := json.Unmarshal([]byte(pmatch.Label.GetValue()), label); err != nil {

				}

				for _, bet := range bets.Bets {
					if bet.MarkUnit == label.Bet {
						if wallet.Chips >= int64(bet.Xplaynow) {
							foundBet = true
							matchBet = bet.MarkUnit
							break
						}
					}
				}

				if foundBet {
					match.Bet = &pb.Bet{
						MarkUnit: matchBet,
					}
					break
				}
			}
		}

		if !foundBet {
			logger.Info("final don't found any matching match")
			return "", presenter.ErrMatchNotFound
		}

		logger.Info("found match with bet %d, id %s", match.Bet.MarkUnit, match.MatchId)

		resMatches.Matches = append(resMatches.Matches, &match)

		response, err := marshaler.Marshal(resMatches)
		if err != nil {
			logger.Error("error marshaling response payload: %v", err.Error())
			return "", presenter.ErrMarshal
		}

		return string(response), nil
	}
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

		// load bet config
		bets, err := LoadBets(request.GameCode, ctx, logger, nk)
		if err != nil {
			return "", presenter.ErrInternalError
		}

		if len(bets.Bets) == 0 {
			return "", presenter.ErrInternalError
		}

		sort.Slice(bets.Bets, func(i, j int) bool {
			return bets.Bets[i].MarkUnit < bets.Bets[j].MarkUnit
		})

		var valid = false
		for _, bet := range bets.Bets {
			if bet.MarkUnit == request.GetMarkUnit() {
				valid = true
				break
			}
		}

		if !valid {
			return "", presenter.ErrBetNotFound
		}

		// No available matches found, create a new one.
		matchID, err := nk.MatchCreate(ctx, request.GameCode, map[string]interface{}{
			"bet":       request.GetMarkUnit(),
			"game_code": request.GameCode,
			"name":      request.Name,
			"password":  request.Password,
		})

		if err != nil {
			logger.Error("error creating match: %v", err)
			return "", presenter.ErrInternalError
		}

		response, err := marshaler.Marshal(&pb.RpcCreateMatchResponse{MatchId: matchID})
		if err != nil {
			logger.Error("error marshaling response payload: %v", err.Error())
			return "", presenter.ErrMarshal
		}

		logger.Info("create match response=", response)

		return string(response), nil
	}
}
