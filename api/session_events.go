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
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	notificationCodeSingleDevice = 101

	streamModeNotification = 0
)

func RegisterSessionEvents(db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	if err := initializer.RegisterEventSessionStart(eventSessionStartFunc(nk, db)); err != nil {
		return err
	}
	if err := initializer.RegisterEventSessionEnd(eventSessionEndFunc(nk, db)); err != nil {
		return err
	}

	initializer.RegisterBeforeSessionLogout(func(ctx context.Context,
		logger runtime.Logger,
		db *sql.DB, nk runtime.NakamaModule,
		in *api.SessionLogoutRequest,
	) (*api.SessionLogoutRequest, error) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			logger.Error("context did not contain user ID.")
			return in, nil
		}
		// Force disconnect the socket for the user's other game client.
		presences, err := nk.StreamUserList(streamModeNotification, userID, "", "", true, true)
		if err != nil || len(presences) == 0 {
			logger.WithField("err", err).Error("nk.StreamUserList error.")
			return in, nil
		}
		for _, presence := range presences {
			sessionId := presence.GetSessionId()
			if err := nk.SessionDisconnect(ctx, sessionId); err != nil {
				logger.WithField("err", err).Error("nk.SessionDisconnect error.")
				return in, nil
			}
			logger.WithField("session id", sessionId).Debug("Session disconnect successful")
		}
		return in, nil
	})
	return nil
}

// Update a user's last online timestamp when they disconnect.
func eventSessionEndFunc(nk runtime.NakamaModule, db *sql.DB) func(context.Context, runtime.Logger, *api.Event) {
	return func(ctx context.Context, logger runtime.Logger, evt *api.Event) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			logger.Error("context did not contain user ID.")
			return
		}
		// saveSecsOnlineNotClaimReward(ctx, logger, nk, db)

		// Restrict the time allowed with the DB operation so we can fail fast in a stampeding herd scenario.
		ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
		query := `UPDATE
					users AS u
				SET
					metadata
						= u.metadata
						|| jsonb_build_object('last_online_time_unix', extract('epoch' FROM now())::BIGINT)
				WHERE	
					id = $1;`

		_, err := db.ExecContext(ctx2, query, userID)
		cancel()
		if err != nil && err != context.DeadlineExceeded {
			logger.WithField("err", err).Error("db.ExecContext last online update error.")
			return
		}
	}
}

// Limit the number of concurrent realtime sessions active for a user to just one.
func eventSessionStartFunc(nk runtime.NakamaModule, db *sql.DB) func(context.Context, runtime.Logger, *api.Event) {
	return func(ctx context.Context, logger runtime.Logger, evt *api.Event) {
		userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if !ok {
			logger.Error("context did not contain user ID.")
			return
		}

		sessionID, ok := ctx.Value(runtime.RUNTIME_CTX_SESSION_ID).(string)
		if !ok {
			logger.Error("context did not contain session ID.")
			return
		}

		ResetUserDailyReward(ctx, logger, nk)

		// Fetch all live presences for this user on their private notification stream.
		{
			presences, err := nk.StreamUserList(streamModeNotification, userID, "", "", true, true)
			if err != nil {
				logger.WithField("err", err).Error("nk.StreamUserList error.")
				return
			}

			notifications := []*runtime.NotificationSend{
				{
					Code: notificationCodeSingleDevice,
					Content: map[string]interface{}{
						"kicked_by": sessionID,
					},
					Persistent: false,
					Sender:     userID,
					Subject:    "Another device is active!",
					UserID:     userID,
				},
			}
			for _, presence := range presences {
				if presence.GetUserId() == userID && presence.GetSessionId() == sessionID {
					// Ignore our current socket connection.
					continue
				}
				ctx2, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				if err := nk.NotificationsSend(ctx2, notifications); err != nil {
					logger.WithField("err", err).Error("nk.NotificationsSend error.")
					continue
				}

			}
		}
		// save login info
		{
			ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
			query := `UPDATE
					users AS u
				SET
					metadata
						= u.metadata
						|| jsonb_build_object('last_login_time_unix', extract('epoch' FROM now())::BIGINT,
																	'last_login_device_id', 'todo',
																	'last_login_ip', 'todo')
				WHERE	
					id = $1;`

			_, err := db.ExecContext(ctx2, query, userID)
			cancel()
			if err != nil && err != context.DeadlineExceeded {
				logger.WithField("err", err).Error("db.ExecContext last online update error.")
				return
			}

		}
	}

}

func ResetUserDailyReward(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		logger.Error("context did not contain user ID.")
		return
	}
	lastClaim, lastClaimObject, err := GetLastDailyRewardObject(ctx, logger, nk)
	if err != nil {
		logger.Error("GetLastDailyRewardObject error %s", err.Error())
		return
	}
	lastClaim.NextClaimUnix = 0
	lastClaim.LastClaimUnix = time.Now().Unix()
	lastClaim.LastSpinNumber = 0
	lastClaim.Streak = 0
	version := ""
	if lastClaimObject != nil {
		version = lastClaimObject.GetVersion()
	}
	SaveLastClaimReward(ctx, nk, logger, lastClaim, version, userID)
}
