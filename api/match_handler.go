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
	"sort"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/ciaolink-game-platform/cgp-lobby-module/api/presenter"
	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
)

const kDefaultMaxSize = 3

type MatchLabel struct {
	Open         int32  `json:"open"`
	Bet          int32  `json:"bet"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Password     string `json:"password"`
	MaxSize      int32  `json:"max_size"`
	MockCodeCard int32  `json:"mock_code_card"`
}

func RpcFindMatch(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		logger.Info("rpc find match: %v", payload)
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		request := &pb.RpcFindMatchRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			return "", presenter.ErrUnmarshal
		}
		if err := checkEnoughChipForBet(ctx, logger, db, nk, userID, int64(request.MarkUnit)); err != nil {
			return "", err
		}

		maxSize := kDefaultMaxSize

		var query string
		if request.WithNonOpen {
			query = fmt.Sprintf("+label.code:%s +label.bet:%d", request.GameCode, request.MarkUnit)
		} else {
			query = fmt.Sprintf("+label.open:>0 +label.code:%s +label.bet:%d", request.GameCode, request.MarkUnit)
		}

		request.MockCodeCard = 0
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
			var label MatchLabel
			err = json.Unmarshal([]byte(match.Label.GetValue()), &label)
			if err != nil {
				logger.Error("unmarshal label error %v", err)
				continue
			}

			logger.Debug("find match size: %v", match.Size)
			if match.Size >= label.MaxSize {
				continue
			}
			resMatches.Matches = append(resMatches.Matches, &pb.Match{
				MatchId:      match.MatchId,
				Size:         match.Size,
				MaxSize:      label.MaxSize, // Get from label
				Name:         label.Name,
				MarkUnit:     label.Bet,
				Open:         label.Open > 0,
				MockCodeCard: label.MockCodeCard,
			})
		}
		if len(resMatches.Matches) <= 0 && request.Create {
			// not found match, auto create match if request need
			arg := map[string]interface{}{
				"bet":      int32(request.MarkUnit),
				"code":     request.GameCode,
				"max_size": int32(2),
				"name":     request.GameCode,
			}
			if request.GetMockCodeCard() > 0 {
				arg["mock_code_card"] = request.GetMockCodeCard()
			}
			// No available matches found, create a new one.
			matchID, err := nk.MatchCreate(ctx, request.GameCode, arg)
			if err != nil {
				logger.Error("error creating match: %v", err)
				return "", presenter.ErrInternalError
			}
			logger.Info("Create new match with arg %v", arg)
			resMatches.Matches = append(resMatches.Matches, &pb.Match{
				MatchId:      matchID,
				Size:         1,
				MaxSize:      arg["max_size"].(int32),
				MarkUnit:     arg["bet"].(int32),
				Open:         true,
				MockCodeCard: request.MockCodeCard,
			})
		}
		//  not found match,
		if len(resMatches.Matches) <= 0 {
			logger.Error("Not found match for user %s", userID)
			return "", presenter.ErrMatchNotFound
		}

		response, err := marshaler.Marshal(resMatches)
		if err != nil {
			logger.Error("error marshaling response payload: %v", err.Error())
			return "", presenter.ErrMarshal
		}

		return string(response), nil
	}
}

func RpcQuickMatch(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		logger.Info("rpc find match: %v", payload)
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		request := &pb.RpcCreateMatchRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			return "", presenter.ErrUnmarshal
		}
		if err := checkEnoughChipForBet(ctx, logger, db, nk, userID, int64(request.MarkUnit)); err != nil {
			return "", err
		}
		maxSize := kDefaultMaxSize
		query := fmt.Sprintf("+label.code:%s +label.open:true", request.GameCode)

		resMatches := &pb.RpcFindMatchResponse{}
		matches, err := nk.MatchList(ctx, -1, true, "", nil, &maxSize, query)
		if err != nil {
			logger.Error("error listing matches: %v", err)
			return "", presenter.ErrInternalError
		}

		logger.Debug("MatchList result %v", matches)
		if len(matches) == 0 {
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
			if err := checkEnoughChipForBet(ctx, logger, db, nk, userID, int64(bets.Bets[0].MarkUnit)); err != nil {
				return "", err
			}
			// No available matches found, create a new one.
			matchID, err := nk.MatchCreate(ctx, request.GameCode, map[string]interface{}{
				"bet":      bets.Bets[0].MarkUnit,
				"code":     request.GameCode,
				"name":     request.Name,
				"password": request.Password,
			})
			if err != nil {
				logger.Error("error creating match: %v", err)
				return "", presenter.ErrInternalError
			}
			resMatches.Matches = append(resMatches.Matches, &pb.Match{
				MatchId:  matchID,
				Size:     1,
				MaxSize:  int32(maxSize),
				Name:     request.Name,
				MarkUnit: bets.Bets[0].MarkUnit,
				Open:     true,
			})
			response, err := marshaler.Marshal(resMatches)
			if err != nil {
				logger.Error("error marshaling response payload: %v", err.Error())
				return "", presenter.ErrMarshal
			}
			return string(response), nil
		}
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
				MatchId:  match.MatchId,
				Size:     match.Size,
				MaxSize:  label.MaxSize, // Get from label
				Name:     label.Name,
				MarkUnit: label.Bet,
			})
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
}

func RpcCreateMatch(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		logger.Info("rpc create match: %v", payload)

		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		request := &pb.RpcCreateMatchRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			logger.Error("unmarshal create match error %v", err)
			return "", presenter.ErrUnmarshal
		}
		if err := checkEnoughChipForBet(ctx, logger, db, nk, userID, int64(request.MarkUnit)); err != nil {
			return "", err
		}
		// No available matches found, create a new one.
		matchID, err := nk.MatchCreate(ctx, request.GameCode, map[string]interface{}{
			"bet":      request.MarkUnit,
			"code":     request.GameCode,
			"name":     request.Name,
			"password": request.Password,
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

func checkEnoughChipForBet(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, userID string, bet int64) error {
	wallet, err := entity.ReadWalletUser(ctx, nk, logger, userID)
	if err != nil {
		logger.Error("read wallet user %s error %s",
			userID, err.Error())
		return presenter.ErrInternalError
	}
	if wallet.Chips <= 0 || wallet.Chips < bet {
		logger.Error("User %s not enough chip [%d] to join game bet [%d]",
			userID, wallet.Chips, bet)
		return presenter.ErrNotEnoughChip
	}
	return nil
}
