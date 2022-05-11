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
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
)

const kDefaultMaxSize = 3

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

		maxSize := kDefaultMaxSize

		var query string
		if request.WithNonOpen {
			query = fmt.Sprintf("+label.code:%s +label.bet:%d", request.GameCode, request.MarkUnit)
		} else {
			query = fmt.Sprintf("+label.open:>0 +label.code:%s +label.bet:%d", request.GameCode, request.MarkUnit)
		}

		logger.Info("match query %v", query)

		resMatches := &pb.RpcFindMatchResponse{}
		matches, err := nk.MatchList(ctx, 10, true, "", nil, &maxSize, query)
		if err != nil {
			logger.Error("error listing matches: %v", err)
			return "", presenter.ErrInternalError
		}

		logger.Debug("find match result %v", matches)
		if len(matches) <= 0 {
			if request.Create {
				// No available matches found, create a new one.
				matchID, err := nk.MatchCreate(ctx, request.GameCode, map[string]interface{}{
					"bet":       request.MarkUnit,
					"game_code": request.GameCode,
				})
				if err != nil {
					logger.Error("error creating match: %v", err)
					return "", presenter.ErrInternalError
				}
				resMatches.Matches = append(resMatches.Matches, &pb.Match{
					MatchId:  matchID,
					Size:     1,
					MaxSize:  int32(maxSize),
					MarkUnit: request.MarkUnit,
					Open:     true,
				})
			}
		} else {
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
					Open:     label.Open > 0,
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

func RpcQuickMatch(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		logger.Info("rpc find match: %v", payload)
		_, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		request := &pb.RpcCreateMatchRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			return "", presenter.ErrUnmarshal
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
			// No available matches found, create a new one.
			matchID, err := nk.MatchCreate(ctx, request.GameCode, map[string]interface{}{
				"bet":       bets.Bets[0].MarkUnit,
				"game_code": request.GameCode,
				"name":      request.Name,
				"password":  request.Password,
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

		_, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}

		request := &pb.RpcCreateMatchRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			logger.Error("unmarshal create match error %v", err)
			return "", presenter.ErrUnmarshal
		}

		// No available matches found, create a new one.
		matchID, err := nk.MatchCreate(ctx, request.GameCode, map[string]interface{}{
			"bet":       request.MarkUnit,
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
