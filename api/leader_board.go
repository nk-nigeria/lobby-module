package api

import (
	"context"
	"database/sql"
	"errors"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nakamaFramework/cgb-lobby-module/api/presenter"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	"github.com/nakamaFramework/cgb-lobby-module/constant"
	pb "github.com/nakamaFramework/cgp-common/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func InitLeaderBoard(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, unmarshaler *protojson.UnmarshalOptions) {
	objectIds := []*runtime.StorageRead{
		{
			Collection: kLobbyCollection,
			Key:        kGameKey,
		},
	}
	objects, err := nk.StorageRead(ctx, objectIds)
	if err != nil {
		logger.Error("Error when read list game, error %s", err.Error())
		return
	}
	if len(objects) == 0 {
		logger.Error("Empty list game in storage ")
		return
	}
	gameListResponse := &pb.GameListResponse{}
	err = unmarshaler.Unmarshal([]byte(objects[0].GetValue()), gameListResponse)
	if err != nil {
		logger.Debug("Can not unmarshaler list game for collection")
		return
	}

	for _, game := range gameListResponse.Games {
		authoritative := true // No client can submit a score directly.
		sort := "desc"
		operator := "incr"
		reset := constant.RESET_SCHEDULER_LEADER_BOARD
		metadata := map[string]interface{}{}
		if err := nk.LeaderboardCreate(ctx, game.Code, authoritative, sort, operator, reset, metadata); err != nil {
			logger.Debug("Can not create leaderboard " + game.Code)
		}
	}
}

func UpdateScoreLeaderBoard(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, leaderBoardRecord *pb.LeaderBoardRecord) {
	accounts, err := nk.AccountsGetId(ctx, []string{leaderBoardRecord.UserId})
	if err != nil || len(accounts) == 0 {
		logger.Error("[UpdateScoreLeaderBoard] AccountsGetId %v", err)
		return
	}
	account := accounts[0]
	if _, err := nk.LeaderboardRecordWrite(ctx, leaderBoardRecord.GameCode, leaderBoardRecord.UserId, account.GetUser().GetUsername(), leaderBoardRecord.Score, 0, map[string]interface{}{}, nil); err != nil {
		logger.Debug("Can not UpdateScoreLeaderBoard %v", leaderBoardRecord)
	} else {
		logger.Info("UpdateScoreLeaderBoard success %v", leaderBoardRecord)
	}
}

func RpcLeaderboardInfo() func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		_, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			return "", errors.New("Missing user ID.")
		}
		req := &pb.LeaderBoardRecord{}
		if err := conf.Unmarshaler.Unmarshal([]byte(payload), req); err != nil {
			logger.Error("Error when unmarshal payload", err.Error())
			return "", presenter.ErrUnmarshal
		}

		if len(req.GameCode) == 0 {
			return "", nil
		}
		list, err := nk.LeaderboardsGetId(ctx, []string{req.GameCode})
		if err != nil {
			logger.WithField("err", err).Error("Error when get leaderboard")
			return "", err
		}
		if len(list) == 0 {
			return "", nil
		}
		board := &pb.LeaderBoardRecord{
			CdResetUnix: int64(list[0].NextReset),
		}
		dataJson, _ := conf.MarshalerDefault.Marshal(board)
		return string(dataJson), nil
	}
}
