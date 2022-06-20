package cgbdb

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"github.com/ciaolink-game-platform/cgp-lobby-module/constant"
	"sort"
	"strconv"
	"strings"

	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//CREATE SEQUENCE user_group_id_seq;
//CREATE TABLE public.user_group (
//  id bigint NOT NULL DEFAULT nextval('user_group_id_seq'),
//  name character varying(256)  NOT NULL,
//  type character varying(128) NOT NULL,
//  data character varying(128) NOT NULL,
//  deleted boolean NOT NULL,
//  create_time timestamp with time zone NOT NULL DEFAULT now(),
//  update_time timestamp with time zone NOT NULL DEFAULT now(),
//	constraint user_group_pk primary key (id),
//	UNIQUE (name)
//);
//ALTER SEQUENCE user_group_id_seq OWNED BY public.user_group.id;
const UserGroupTableName = "user_group"

func AddUserGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, userGroup *pb.UserGroup) error {
	if userGroup == nil || userGroup.Name == "" || userGroup.Type == "" {
		return status.Error(codes.InvalidArgument, "Error add user group.")
	}
	query := "INSERT INTO " + UserGroupTableName + " (name, type, data, deleted, create_time, update_time) VALUES ($1, $2, $3, false, now(), now())"
	result, err := db.ExecContext(ctx, query, userGroup.Name, userGroup.Type, userGroup.Data)
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

func GetUserGroupById(ctx context.Context, logger runtime.Logger, db *sql.DB, id int64) (*pb.UserGroup, error) {
	if id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Id is empty")
	}
	query := "SELECT id, name, type, data FROM " + UserGroupTableName + " WHERE id=$1 AND deleted = false"
	var dbID int64
	var dbName, dbType, dbData string
	err := db.QueryRowContext(ctx, query, id).Scan(&dbID, &dbName, &dbType, &dbData)
	if err != nil {
		logger.Error("Query user_group by id %, error %s", id, err.Error())
		return nil, status.Error(codes.Internal, "Query user_group error")
	}
	userGroup := pb.UserGroup{
		Id:   dbID,
		Name: dbName,
		Type: dbType,
		Data: dbData,
	}
	return &userGroup, nil
}

func UpdateUserGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, id int64, userGroup *pb.UserGroup) (*pb.UserGroup, error) {
	oldUserGroup, err := GetUserGroupById(ctx, logger, db, id)
	if err != nil || oldUserGroup.Name == "" {
		return nil, err
	}
	query := "UPDATE " + UserGroupTableName + " SET name=$1, data=$2 WHERE id=$3"
	result, err := db.ExecContext(ctx, query, userGroup.Name, userGroup.Data, id)
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
	return oldUserGroup, nil
}

func DeleteUserGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, id int64) error {
	oldUserGroup, err := GetUserGroupById(ctx, logger, db, id)
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

func GetListUserIdsByUserGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, id int64) ([]string, error) {
	userGroup, err := GetUserGroupById(ctx, logger, db, id)
	if err != nil || userGroup.Name == "" {
		return nil, err
	}
	condition := ""
	params := make([]interface{}, 0)
	typeUG := constant.UserGroupType(userGroup.Type)
	operator := ""
	param := ""

	if strings.Contains(userGroup.Data, "=") {
		operator = userGroup.Data[:2]
		param = strings.Trim(userGroup.Data[2:len(userGroup.Data)], " ")
	} else {
		operator = string(userGroup.Data[0])
		param = strings.Trim(userGroup.Data[1:len(userGroup.Data)], " ")
	}

	if typeUG == constant.UserGroupType_WalletChips {
		condition = fmt.Sprintf(" WHERE (wallet->>'chips')::bigint %s $1", operator)
	} else if typeUG == constant.UserGroupType_WalletChipsInbank {
		condition = fmt.Sprintf(" WHERE (wallet->>'chipsInBank')::bigint %s $1", operator)
	} else if typeUG == constant.UserGroupType_Level {
		condition = fmt.Sprintf(" WHERE (metadata->>'level')::bigint %s $1", operator)
	} else if typeUG == constant.UserGroupType_VipLevel {
		condition = fmt.Sprintf(" WHERE (metadata->>'vip_level')::bigint %s $1", operator)
	}
	paramN, _ := strconv.ParseInt(param, 10, 64)
	params = append(params, paramN)
	logger.Info("FetchUserIDWithCondition %v %v", condition, params)
	return FetchUserIDWithCondition(ctx, db, condition, params)
}

func GetListUserGroup(ctx context.Context, logger runtime.Logger, db *sql.DB, limit int64, cursor string) (*pb.ListUserGroup, error) {
	var incomingCursor = &entity.UserGroupListCursor{}
	if cursor != "" {
		cb, err := base64.URLEncoding.DecodeString(cursor)
		if err != nil {
			return nil, ErrWalletLedgerInvalidCursor
		}
		if err := gob.NewDecoder(bytes.NewReader(cb)).Decode(incomingCursor); err != nil {
			return nil, ErrWalletLedgerInvalidCursor
		}

		logger.Info("GetListUserGroup with cusor %d, create time %s ",
			incomingCursor.Id,
			incomingCursor.CreateTime.String())
	}

	if limit <= 0 {
		limit = 100
	}
	if incomingCursor.Id < 0 {
		incomingCursor.Id = 0
	}

	var rows *sql.Rows
	var err error

	params := make([]interface{}, 0)
	query := ""

	if incomingCursor.Id > 0 {
		if incomingCursor.IsNext {
			query += " WHERE id < $1 AND deleted = false order by id desc "
		} else {
			query += " WHERE id > $1 AND deleted = false order by id asc"
		}
		params = append(params, incomingCursor.Id)
		query += "  limit $2"
		params = append(params, limit)
	} else {
		query += " WHERE deleted = false  order by id desc limit $1"
		params = append(params, limit)
	}
	queryRow := "SELECT id, name, type, data FROM " +
		UserGroupTableName + query
	rows, err = db.QueryContext(ctx, queryRow, params...)
	if err != nil {
		logger.Error("Query lists user group, error %s", err.Error())
		return nil, status.Error(codes.Internal, "Query lists user group")
	}
	ml := make([]*pb.UserGroup, 0)
	var dbID int64
	var dbName, dbType, dbData string
	for rows.Next() {
		rows.Scan(&dbID, &dbName, &dbType, &dbData)
		userGroup := pb.UserGroup{
			Id:   dbID,
			Name: dbName,
			Type: dbType,
			Data: dbData,
		}
		ml = append(ml, &userGroup)
	}
	sort.Slice(ml, func(i, j int) bool {
		return ml[i].Id > ml[j].Id
	})
	var total int64 = incomingCursor.Total
	if total <= 0 {
		queryTotal := "Select count(*) as total FROM " + UserGroupTableName +
			strings.ReplaceAll(query, "order by id desc", "")

		_ = db.QueryRowContext(ctx, queryTotal, params...).Scan(&total)
	}
	var nextCursor *entity.UserGroupListCursor
	var prevCursor *entity.UserGroupListCursor
	if len(ml) > 0 {
		if len(ml)+int(incomingCursor.Offset) < int(total) {
			nextCursor = &entity.UserGroupListCursor{
				Id:     ml[len(ml)-1].Id,
				IsNext: true,
				Offset: incomingCursor.Offset + int64(len(ml)),
				Total:  total,
			}
		}

		prevOffset := incomingCursor.Offset - int64(len(ml))
		if len(ml)+int(incomingCursor.Offset) >= int(total) {
			prevOffset = total - int64(len(ml)) - limit
		}
		if prevOffset < 0 {
			prevOffset = 0
		}
		prevCursor = &entity.UserGroupListCursor{
			Id:     ml[0].Id,
			IsNext: false,
			Offset: prevOffset,
			Total:  total,
		}
	}

	var nextCursorStr string
	if nextCursor != nil {
		cursorBuf := new(bytes.Buffer)
		if err := gob.NewEncoder(cursorBuf).Encode(nextCursor); err != nil {
			logger.Error("Error creating wallet ledger list cursor", zap.Error(err))
			return nil, err
		}
		nextCursorStr = base64.URLEncoding.EncodeToString(cursorBuf.Bytes())
	}

	var prevCursorStr string
	if prevCursor != nil {
		cursorBuf := new(bytes.Buffer)
		if err := gob.NewEncoder(cursorBuf).Encode(prevCursor); err != nil {
			logger.Error("Error creating wallet ledger list cursor", zap.Error(err))
			return nil, err
		}
		prevCursorStr = base64.URLEncoding.EncodeToString(cursorBuf.Bytes())
	}

	return &pb.ListUserGroup{
		UserGroups: ml,
		NextCusor:  nextCursorStr,
		PrevCusor:  prevCursorStr,
		Total:      total,
		Offset:     incomingCursor.Offset,
		Limit:      limit,
	}, nil
}
