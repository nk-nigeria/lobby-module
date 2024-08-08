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
	"strings"
	"sync"
	"time"

	"github.com/nakamaFramework/cgp-common/define"
	"github.com/nakamaFramework/cgp-common/lib"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	"github.com/nakamaFramework/cgb-lobby-module/cgbdb"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	"github.com/nakamaFramework/cgb-lobby-module/entity"
	pb "github.com/nakamaFramework/cgp-common/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

const kDefaultMaxSize = 3

var GameStateDurationByGame = make(map[define.GameName]lib.GameStateDuration)

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
		var counter int64 = time.Now().Unix()
		var mt sync.Mutex
		return func() string {
			mt.Lock()
			counter += 1
			newVal := counter
			mt.Unlock()
			str := fmt.Sprintf("%05d", newVal)
			return str[len(str)-5:]
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
		if _, err := checkEnoughChipForBet(ctx, logger, db, nk, userID, request.GameCode, int64(request.MarkUnit), false); err != nil {
			return "", err
		}

		maxSize := define.GetMaxSizeByGame(define.GameName(request.GameCode))

		queryBuilder := strings.Builder{}
		queryBuilder.WriteString(fmt.Sprintf("+label.name:%s ", request.GameCode))
		if request.MarkUnit > 0 {
			queryBuilder.WriteString(fmt.Sprintf("+label.markUnit:%d", request.MarkUnit))
		}
		if len(request.TableId) > 0 {
			queryBuilder.WriteString(fmt.Sprintf("+label.tableId:%s ", request.TableId))
		}
		if request.WithNonOpen {
			queryBuilder.WriteString(fmt.Sprintf("+label.open:>0"))
		}

		// request.MockCodeCard = 0
		if request.MockCodeCard > 0 {
			// query += fmt.Sprintf(" +label.mock_code_card:%d", request.MockCodeCard)
			queryBuilder.WriteString(fmt.Sprintf("+label.mock_code_card:%d ", request.MockCodeCard))
		}
		query := queryBuilder.String()

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
			resMatches.Matches, err = createMatch(ctx, logger, db, nk, false, &pb.RpcCreateMatchRequest{
				GameCode: request.GameCode,
				MarkUnit: request.MarkUnit,
				// MaxSize:  int64(maxSize),
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
			match.Open = len(match.Password) == 0
			match.Password = ""
			match.GameStateDuration = GetGameStateDurationByGameName(define.GameName(match.Name)).ToPb()
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
	// if err := checkEnoughChipForBet(ctx, logger, db, nk, userID, request.GameCode, int64(request.MarkUnit), true); err != nil {
	// 	logger.WithField("user", userID).WithField("min chip", request.MarkUnit).WithField("err", err).Error("not enough chip for bet")
	// 	return "", presenter.ErrNotEnoughChip
	// }
	bestBet, err := findMaxBetForUser(ctx, logger, db, nk, userID, request.GameCode, false)
	if err != nil {
		logger.WithField("user", userID).WithField("min chip", request.MarkUnit).WithField("err", err).Error("not enough chip for bet")
		return "", err
	}

	resMatches := &pb.RpcFindMatchResponse{}
	var matches []*api.Match
	// var err error
	if define.IsAllowJoinInGameOnProgress(request.GameCode) {
		maxSize := define.GetMaxSizeByGame(define.GameName(request.GameCode))
		query := fmt.Sprintf("+label.code:%s +label.open:1", request.GameCode)
		if bestBet.MarkUnit > 0 {
			query += fmt.Sprintf(" +label.markUnit:%d", bestBet.MarkUnit)
		}
		matches, err = nk.MatchList(ctx, 100, true, "", nil, &maxSize, query)
		if err != nil {
			logger.Error("error listing matches: %v", err)
			return "", presenter.ErrInternalError
		}

		logger.Debug("MatchList result %v", matches)
	}

	if len(matches) == 0 {
		resMatches.Matches, err = createMatch(ctx, logger, db, nk, true, request)
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
		matchs, err := createMatch(ctx, logger, db, nk, false, request)
		if err != nil {
			logger.WithField("err", err).Error("error creating match")
			return "", err
		}
		logger.WithField("matchs", matchs).Info("created match")
		if len(matchs) == 0 {
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

func RpcInfoMatch(marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		_, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", presenter.ErrNoUserIdFound
		}
		request := &pb.MatchInfoRequest{}
		if err := unmarshaler.Unmarshal([]byte(payload), request); err != nil {
			logger.Error("unmarshal create match error %v", err)
			return "", presenter.ErrUnmarshal
		}
		match, err := nk.MatchGet(ctx, request.MatchId)
		if err != nil {
			logger.WithField("err", err).WithField("match_id", request.MatchId).Error("error get match")
			return "", presenter.ErrMatchNotFound
		}
		if match == nil {
			return "", presenter.ErrMatchNotFound
		}
		matchInfo := &pb.Match{}
		err = conf.Unmarshaler.Unmarshal([]byte(match.Label.GetValue()), matchInfo)
		if err != nil {
			logger.Error("unmarshal label error %v", err)
			return "", presenter.ErrUnmarshal
		}
		if request.QueryUser {
			userUuids := make([]string, 0)
			for _, profile := range matchInfo.Profiles {
				userUuids = append(userUuids, profile.GetUserId())
			}
			accounts, err := cgbdb.GetProfileUsers(ctx, db, userUuids...)
			if err != nil {
				logger.WithField("err", err).Error("get account failed")
				return "", presenter.ErrInternalError
			}
			accountByUuid := accounts.ToMap()
			// matchInfo.Profiles = make([]*pb.SimpleProfile, 0)
			for idx, profile := range matchInfo.Profiles {
				v := accountByUuid[profile.UserId]
				// matchInfo.Profiles = append(matchInfo.Profiles, profile)
				matchInfo.Profiles[idx] = v
			}
		}
		matchInfo.GameStateDuration = GetGameStateDurationByGameName(define.GameName(matchInfo.Name)).ToPb()

		response, err := conf.MarshalerDefault.Marshal(matchInfo)
		if err != nil {
			logger.Error("error marshaling response payload: %v", err.Error())
			return "", presenter.ErrMarshal
		}
		return string(response), nil
	}
}

func createMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, quickJoin bool, request *pb.RpcCreateMatchRequest) ([]*pb.Match, error) {
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
		matchInfo.MaxSize = int32(define.GetMaxSizeByGame(define.GameName(matchInfo.Name)))
	}
	// matchInfo.MaxSize = 4
	if matchInfo.NumBot <= 0 {
		matchInfo.NumBot = 1
	}
	// bets, err := LoadBets(ctx, logger, db, nk, request.GameCode)
	// if err != nil {
	// 	logger.WithField("err", err).Error("load bets failed")
	// 	return nil, presenter.ErrInternalError
	// }

	// if request.MarkUnit <= 0 {
	// 	bestBet, err := findMaxBetForUser(ctx, logger, db, nk, userID, request.GameCode, true)
	// 	if err != nil {
	// 		logger.WithField("user", userID).WithField("gameCode", request.GameCode).WithField("err", err).Error("not enough chip for bet")
	// 		return nil, err
	// 	}
	// 	matchInfo.MarkUnit = int32(bestBet.MarkUnit)
	// }
	if IsNeedCheckBets(request.GameCode) {
		bets, err := loadBetsForUser(ctx, logger, db, nk, request.GameCode, quickJoin, userID)
		if err != nil {
			logger.WithField("user", userID).WithField("gameCode", request.GameCode).WithField("err", err).Error("load bets failed")
			return nil, err
		}
		if bets.BestChoice == nil || !bets.BestChoice.Enable {
			logger.WithField("user", userID).WithField("gameCode", request.GameCode).WithField("err", err).Error("list bets is empty")
			return nil, presenter.ErrNotEnoughChip
		}
		if matchInfo.MarkUnit == 0 {
			matchInfo.MarkUnit = int32(bets.BestChoice.MarkUnit)
		}
		// check mark unit allowed
		allowed := false
		for _, bet := range bets.Bets {
			if int32(bet.MarkUnit) == matchInfo.MarkUnit {
				allowed = bet.Enable
				matchInfo.Bet = bet
				break
			}
		}
		if !allowed {
			return nil, presenter.ErrInvalidInput
		}
	}
	// No available matches found, create a new one.
	arg := make(map[string]any)
	matchInfo.TableId = GetTableId()
	data, _ := conf.MarshalerDefault.Marshal(matchInfo)
	arg["data"] = string(data)
	matchID, err := nk.MatchCreate(ctx, request.GameCode, arg)
	if err != nil {
		logger.WithField("data", string(data)).Error("error creating match: %v", err)
		return nil, presenter.ErrInternalError
	}
	matchInfo.MatchId = matchID
	matchInfo.NumBot = 0
	matchInfo.MockCodeCard = 0
	matchInfo.Open = len(matchInfo.Password) == 0
	matchInfo.Password = ""
	matchInfo.GameStateDuration = GetGameStateDurationByGameName(define.GameName(matchInfo.Name)).ToPb()
	resMatches := &pb.RpcFindMatchResponse{}
	resMatches.Matches = append(resMatches.Matches, matchInfo)
	return resMatches.Matches, nil
}
func checkEnoughChipForBet(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, userID string, gameCode string, betWantCheck int64, quickJoin bool) (*pb.Bet, error) {
	bets, err := LoadBets(ctx, logger, db, nk, gameCode)
	if err != nil {
		return nil, presenter.ErrInternalError
	}
	if len(bets) == 0 {
		return &pb.Bet{
			MarkUnit:  float32(betWantCheck),
			AgJoin:    betWantCheck,
			AgPlayNow: betWantCheck,
			AgLeave:   betWantCheck,
		}, nil
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
		return nil, presenter.ErrBetNotFound
	}
	minChipRequire := bet.AGJoin
	if quickJoin {
		minChipRequire = bet.AGPlaynow
	}
	wallet, err := entity.ReadWalletUser(ctx, nk, logger, userID)
	if err != nil {
		logger.Error("read wallet user %s error %s",
			userID, err.Error())
		return nil, presenter.ErrInternalError
	}
	if wallet.Chips <= 0 || wallet.Chips < int64(minChipRequire) {
		logger.Error("User %s not enough chip [%d] to join game bet [%d]",
			userID, wallet.Chips, bet)
		return nil, presenter.ErrNotEnoughChip
	}
	return bet.ToPb(), nil
}

func findMaxBetForUser(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, userID string, gameCode string, quickJoin bool) (entity.Bet, error) {
	// bets, err := LoadBets(ctx, logger, db, nk, gameCode)
	// if err != nil {
	// 	return entity.Bet{}, presenter.ErrInternalError
	// }
	// if len(bets) == 0 {
	// 	return entity.Bet{}, nil
	// }

	// // sort desc by mark unit
	// sort.Slice(bets, func(i, j int) bool {
	// 	return bets[i].MarkUnit > bets[j].MarkUnit
	// })
	// wallet, err := entity.ReadWalletUser(ctx, nk, logger, userID)
	// if err != nil {
	// 	logger.Error("read wallet user %s error %s",
	// 		userID, err.Error())
	// 	return entity.Bet{}, presenter.ErrInternalError
	// }
	// if wallet.Chips <= 0 {
	// 	return entity.Bet{}, presenter.ErrNotEnoughChip
	// }
	// for _, bet := range bets {
	// 	minChipRequire := bet.AGJoin
	// 	if quickJoin {
	// 		minChipRequire = bet.AGPlaynow
	// 	}
	// 	if wallet.Chips < int64(minChipRequire) {
	// 		continue
	// 	}
	// 	return bet, nil
	// }
	// return entity.Bet{}, presenter.ErrNotEnoughChip
	// quickJoin := true
	bets, err := loadBetsForUser(ctx, logger, db, nk, gameCode, quickJoin, userID)
	if err != nil {
		return entity.Bet{}, err
	}
	if bets == nil || bets.BestChoice == nil || !bets.BestChoice.Enable {
		return entity.Bet{}, nil
	}
	return *entity.PbBetToBet(bets.BestChoice), nil
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
	markUnitSuggest := int64(0)
	req.MarkUnit = 0
	if len(profile.PlayingMatch.Code) != 0 {
		gameCode = profile.PlayingMatch.Code
		// req.MarkUnit = int32(profile.PlayingMatch.Mcb)
		markUnitSuggest = profile.PlayingMatch.Mcb
		req.LastBet = profile.PlayingMatch.Bet
	}
	bets, err := loadBetsForUser(ctx, logger, db, nk, gameCode, true, userID)
	if err != nil {
		logger.WithField("user id", userID).WithField("err", err).Error("load bets failed")
		return "", err
	}
	if bets != nil {
		for _, bet := range bets.Bets {
			if !bet.Enable {
				continue
			}
			if markUnitSuggest == bet.AgPlayNow {
				req.MarkUnit = int32(markUnitSuggest)
				break
			}
		}
		if req.MarkUnit == 0 && bets.BestChoice != nil && bets.BestChoice.Enable {
			req.MarkUnit = int32(bets.BestChoice.MarkUnit)
		}
	}

	req.GameCode = gameCode
	return RpcQuickMatch(ctx, logger, db, nk, req.String())
}

func IsNeedCheckBets(gameCode string) bool {
	if define.IsSlotGame(define.GameName(gameCode)) {
		return false
	}
	return true
}

func GetGameStateDurationByGameName(gameName define.GameName) lib.GameStateDuration {
	v, exist := GameStateDurationByGame[gameName]
	if !exist {
		v = lib.GetGameStateDurationByGameCode(gameName)
		GameStateDurationByGame[gameName] = v
	}
	return v
}
