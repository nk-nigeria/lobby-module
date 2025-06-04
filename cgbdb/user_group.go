package cgbdb

import (
	"context"
	"database/sql"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/heroiclabs/nakama-common/runtime"
	pb "github.com/nakama-nigeria/cgp-common/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CREATE SEQUENCE user_group_id_seq;
// CREATE TABLE public.user_group (
//
//	 id bigint NOT NULL DEFAULT nextval('user_group_id_seq'),
//	 name character varying(256)  NOT NULL,
//	 deleted boolean NOT NULL,
//	 create_time timestamp with time zone NOT NULL DEFAULT now(),
//	 update_time timestamp with time zone NOT NULL DEFAULT now(),
//		constraint user_group_pk primary key (id),
//		UNIQUE (name)
//
// );
// ALTER SEQUENCE user_group_id_seq OWNED BY public.user_group.id;
const UserGroupTableName = "user_group"

func AddUserGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, userGroup *pb.UserGroup, marshaler *proto.MarshalOptions) error {
	if userGroup == nil || userGroup.Name == "" {
		return status.Error(codes.InvalidArgument, "Error add user group.")
	}
	query := "INSERT INTO " + UserGroupTableName + " (name, type, condition, deleted, create_time, update_time) VALUES ($1, $2, $3, false, now(), now())"
	result, err := db.ExecContext(ctx, query, userGroup.Name, "", "")
	if err != nil {
		logger.Error("Add new usergroup, name: %s, error %s",
			userGroup.Name, err.Error())
		if strings.Contains(err.Error(), "duplicate key value") {
			return status.Error(codes.AlreadyExists, "UserGroup name is exists")
		}
		return status.Error(codes.Internal, "Error add user group.")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert new usergroup, name: %s",
			userGroup.Name)
		return status.Error(codes.Internal, "Error add user group.")
	}
	return nil
}

func GetUserGroupById(ctx context.Context, logger runtime.Logger, db *sql.DB, unmarshaler *proto.UnmarshalOptions, id int64) (*pb.UserGroup, error) {
	if id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Id is empty")
	}
	query := "SELECT id, name, type, condition FROM " + UserGroupTableName + " WHERE id=$1 AND deleted = false"
	var dbID int64
	var dbName, dbType string
	var dbCondition []byte
	err := db.QueryRowContext(ctx, query, id).Scan(&dbID, &dbName, &dbType, &dbCondition)
	if err != nil {
		logger.Error("Query user_group by id %, error %s", id, err.Error())
		return nil, status.Error(codes.Internal, "Query user_group error")
	}

	userGroup := pb.UserGroup{
		Id:   dbID,
		Name: dbName,
	}
	return &userGroup, nil
}

func UpdateUserGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, marshaler *proto.MarshalOptions, unmarshaler *proto.UnmarshalOptions, id int64, userGroup *pb.UserGroup) (*pb.UserGroup, error) {
	oldUserGroup, err := GetUserGroupById(ctx, logger, db, unmarshaler, id)
	if err != nil || oldUserGroup.Name == "" {
		return nil, err
	}
	query := "UPDATE " + UserGroupTableName + " SET name=$1 WHERE id=$3"
	result, err := db.ExecContext(ctx, query, userGroup.Name, id)
	if err != nil {
		logger.Error("Update user group id %d, user %s, error %s", id, err.Error())
		return nil, status.Error(codes.Internal, "Update user group error")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not update user group")
		return nil, status.Error(codes.Internal, "Error Update user group")
	}
	oldUserGroup.Name = userGroup.Name
	return oldUserGroup, nil
}

func DeleteUserGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, unmarshaler *proto.UnmarshalOptions, id int64) error {
	oldUserGroup, err := GetUserGroupById(ctx, logger, db, unmarshaler, id)
	if err != nil || oldUserGroup.Name == "" {
		return err
	}
	query := "UPDATE " + UserGroupTableName + " SET deleted=true WHERE id=$1"
	result, err := db.ExecContext(ctx, query, id)
	if err != nil {
		logger.Error("Delete user group id %d, error %s", id, err.Error())
		return status.Error(codes.Internal, "Delete user group")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did delete user group")
		return status.Error(codes.Internal, "Error Delete user group")
	}
	return nil
}
