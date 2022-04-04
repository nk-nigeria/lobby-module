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

type MatchLabel struct {
	Open              int32   `json:"open"`
	LastOpenValueNoti int32   `json:"-"` // using for check has noti new state of open
	Bet               *pb.Bet `json:"bet"`
	Code              string  `json:"code"`
	Name              string  `json:"name"`
	Password          string  `json:"password"`
	MaxSize           int32   `json:"max_size"`
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

		maxSize := 1
		query := fmt.Sprintf("+label.game_code:%s +label.bet.mark_unit:%d", request.GameCode, request.Bet.MarkUnit)

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
					Bet:     label.Bet,
				})
			}
		} else {
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
		maxSize := 2
		query := fmt.Sprintf("+label.game_code:%s +label.bet.mark_unit:%d", request.GameCode, request.Bet.MarkUnit)

		resMatches := &pb.RpcFindMatchResponse{}
		matches, err := nk.MatchList(ctx, 10, true, "", nil, &maxSize, query)
		if err != nil {
			logger.Error("error listing matches: %v", err)
			return "", presenter.ErrInternalError
		}

		logger.Debug("MatchList result %v", matches)
		if len(matches) == 0 {
			bets, err := entity.LoadBets(request.GameCode, ctx, logger, nk, unmarshaler)
			if err != nil {
				return "", presenter.ErrInternalError
			}
			if len(bets.Bets) == 0 {
				return "", nil
			}
			sort.Slice(bets.Bets, func(i, j int) bool {
				return bets.Bets[i].GetMarkUnit() < bets.Bets[j].GetMarkUnit()
			})
			// No available matches found, create a new one.
			matchID, err := nk.MatchCreate(ctx, request.GameCode, map[string]interface{}{
				"bet":       bets.Bets[0],
				"game_code": request.GameCode,
				"name":      request.Name,
				"password":  request.Password,
			})
			if err != nil {
				logger.Error("error creating match: %v", err)
				return "", presenter.ErrInternalError
			}
			resMatches.Matches = append(resMatches.Matches, &pb.Match{
				MatchId: matchID,
				Size:    0,
				MaxSize: int32(maxSize),
				Name:    request.Name,
				Bet:     bets.Bets[0],
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
				MatchId: match.MatchId,
				Size:    match.Size,
				MaxSize: label.MaxSize, // Get from label
				Name:    label.Name,
				Bet:     label.Bet,
			})
		}

		sort.Slice(resMatches.Matches, func(i, j int) bool {
			r := resMatches.Matches[i].Bet.MarkUnit < resMatches.Matches[j].Bet.MarkUnit
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
			"bet":       request.Bet,
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
