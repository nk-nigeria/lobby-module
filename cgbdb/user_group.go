package cgbdb

import (
	"context"
	"database/sql"
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"strings"

	pb "github.com/ciaolink-game-platform/cgp-common/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CREATE SEQUENCE user_group_id_seq;
// CREATE TABLE public.user_group (
//
//	 id bigint NOT NULL DEFAULT nextval('user_group_id_seq'),
//	 name character varying(256)  NOT NULL,
//	 type character varying(128) NOT NULL,
//	 condition jsonb NOT NULL,
//	 deleted boolean NOT NULL,
//	 create_time timestamp with time zone NOT NULL DEFAULT now(),
//	 update_time timestamp with time zone NOT NULL DEFAULT now(),
//		constraint user_group_pk primary key (id),
//		UNIQUE (name)
//
// );
// ALTER SEQUENCE user_group_id_seq OWNED BY public.user_group.id;
const UserGroupTableName = "user_group"

func parseOperator(condition string) (string, string) {
	if len(strings.Trim(condition, " ")) == 0 {
		return "", ""
	}

	operator := ""
	value := ""
	if strings.Contains(condition, "=") {
		operator = condition[:2]
		value = strings.Trim(condition[2:len(condition)], " ")
	} else {
		operator = string(condition[0])
		value = strings.Trim(condition[1:len(condition)], " ")
	}
	return operator, value
}

func AddUserGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, userGroup *pb.UserGroup, marshaler *protojson.MarshalOptions) error {
	if userGroup == nil || userGroup.Name == "" || userGroup.Type == "" {
		return status.Error(codes.InvalidArgument, "Error add user group.")
	}
	operator, value := parseOperator(strings.Trim(userGroup.Data, " "))
	userGroup.Condition = &pb.UserGroupCondition{
		Operator: operator,
		Value:    value,
	}
	conditionStr, _ := marshaler.Marshal(userGroup.Condition)
	query := "INSERT INTO " + UserGroupTableName + " (name, type, condition, deleted, create_time, update_time) VALUES ($1, $2, $3, false, now(), now())"
	result, err := db.ExecContext(ctx, query, userGroup.Name, userGroup.Type, conditionStr)
	if err != nil {
		logger.Error("Add new usergroup, name: %s, type: %s, data: %s, error %s",
			userGroup.Name, userGroup.Type, userGroup.Data, err.Error())
		if strings.Contains(err.Error(), "duplicate key value") {
			return status.Error(codes.AlreadyExists, "UserGroup name is exists")
		}
		return status.Error(codes.Internal, "Error add user group.")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert new usergroup, name: %s, type: %s, data: %s",
			userGroup.Name, userGroup.Type, userGroup.Data)
		return status.Error(codes.Internal, "Error add user group.")
	}
	return nil
}

func GetUserGroupById(ctx context.Context, logger runtime.Logger, db *sql.DB, unmarshaler *protojson.UnmarshalOptions, id int64) (*pb.UserGroup, error) {
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
	var condition = &pb.UserGroupCondition{}
	err = unmarshaler.Unmarshal(dbCondition, condition)
	if err != nil {
		logger.Error("Unmarshal dbCondition error %s", err.Error())
		return nil, status.Error(codes.Internal, "Query user_group error")
	}
	userGroup := pb.UserGroup{
		Id:        dbID,
		Name:      dbName,
		Type:      dbType,
		Data:      fmt.Sprintf("%s%s", condition.Operator, condition.Value),
		Condition: condition,
	}
	return &userGroup, nil
}

func UpdateUserGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions, id int64, userGroup *pb.UserGroup) (*pb.UserGroup, error) {
	oldUserGroup, err := GetUserGroupById(ctx, logger, db, unmarshaler, id)
	if err != nil || oldUserGroup.Name == "" {
		return nil, err
	}
	operator, value := parseOperator(strings.Trim(userGroup.Data, " "))
	userGroup.Condition = &pb.UserGroupCondition{
		Operator: operator,
		Value:    value,
	}
	conditionStr, _ := marshaler.Marshal(userGroup.Condition)
	query := "UPDATE " + UserGroupTableName + " SET name=$1, condition=$2 WHERE id=$3"
	result, err := db.ExecContext(ctx, query, userGroup.Name, conditionStr, id)
	if err != nil {
		logger.Error("Update user group id %d, user %s, error %s", id, err.Error())
		return nil, status.Error(codes.Internal, "Update user group error")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not update user group")
		return nil, status.Error(codes.Internal, "Error Update user group")
	}
	oldUserGroup.Name = userGroup.Name
	oldUserGroup.Data = userGroup.Data
	oldUserGroup.Condition = userGroup.Condition
	return oldUserGroup, nil
}

func DeleteUserGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, unmarshaler *protojson.UnmarshalOptions, id int64) error {
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
