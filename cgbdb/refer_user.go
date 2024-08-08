package cgbdb

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgtype"
	"github.com/nakamaFramework/cgb-lobby-module/conf"
	pb "github.com/nakamaFramework/cgp-common/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/heroiclabs/nakama-common/runtime"
)

const ReferUserTableName = "referuser"

// CREATE TABLE
//   public.referuser (
//      id bigint NOT NULL,
// 	   user_invitor character varying(128) NOT NULL,
// 	   user_invitee character varying(128) NOT NULL,
// 		UNIQUE(user_invitee),
//     create_time timestamp
//     with
//       time zone NOT NULL DEFAULT now(),
//       update_time timestamp
//     with
//       time zone NOT NULL DEFAULT now()
//   );

// ALTER TABLE
//   public.referuser
// ADD
//   CONSTRAINT referuser_pkey PRIMARY KEY (id)

func AddUserRefer(ctx context.Context, logger runtime.Logger, db *sql.DB, userRefer *pb.ReferUser) (int64, error) {
	userRefer.Id = conf.SnowlakeNode.Generate().Int64()
	query := "INSERT INTO " + ReferUserTableName +
		" (id, user_invitor, user_invitee, create_time, update_time) VALUES ($1, $2, $3, now(), now())"
	result, err := db.ExecContext(ctx, query,
		userRefer.GetId(), userRefer.GetUserInvitor(), userRefer.GetUserInvitee())
	if err != nil {
		logger.Error("Error when add new refer user, user invitor: %s, user invitee: %s error %s",
			userRefer.GetUserInvitor(), userRefer.GetUserInvitee(), err.Error())
		return 0, status.Error(codes.Internal, "Error add use refer.")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert new user refer, user invitor: %s, user invitee: %s",
			userRefer.GetUserInvitor(), userRefer.GetUserInvitee())
		return 0, status.Error(codes.Internal, "Error add use refer.")
	}
	return userRefer.Id, nil
}

func ListUserInvitedByUserId(ctx context.Context, logger runtime.Logger, db *sql.DB, userId string) ([]*pb.ReferUser, error) {
	if userId == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid user")
	}
	query := "Select id, user_invitor, user_invitee, create_time FROM " + ReferUserTableName + " WHERE user_invitor=$1"
	rows, err := db.QueryContext(ctx, query, userId)
	if err != nil {
		logger.Error("Query list user invited by user %s error ", userId, err.Error())
		return nil, status.Error(codes.Internal, "Query list user invited error")
	}
	var dbID int64
	var dbUserInvitorId, dbUserInviteeId string
	var dbCreateTime pgtype.Timestamptz
	ml := make([]*pb.ReferUser, 0)
	for rows.Next() {
		rows.Scan(&dbID, &dbUserInvitorId, &dbUserInviteeId, &dbCreateTime)
		referUser := &pb.ReferUser{
			Id:             dbID,
			UserInvitor:    dbUserInvitorId,
			UserInvitee:    dbUserInviteeId,
			CreateTimeUnix: dbCreateTime.Time.Unix(),
		}
		ml = append(ml, referUser)
	}
	return ml, err
}

func GetAllUserHasReferLeastOneUser(ctx context.Context, logger runtime.Logger, db *sql.DB, timeCreated *time.Time) ([]*pb.ReferUser, error) {
	query := "Select DISTINCT user_invitor FROM " + ReferUserTableName
	args := make([]interface{}, 0)
	if timeCreated != nil {
		query += " WHERE create_time <=$1"
		args = append(args, timeCreated)
	}
	var dbUserId string
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Error("GetAllUserHasReferLeastOneUser error %s", err.Error())
		return nil, nil
	}
	ml := make([]*pb.ReferUser, 0)
	for rows.Next() {
		if rows.Scan(&dbUserId) == nil {
			r := &pb.ReferUser{
				UserInvitor: dbUserId,
			}
			ml = append(ml, r)
		}
	}
	return ml, nil
}
