package cgbdb

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"strings"
	"time"

	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//CREATE SEQUENCE in_app_message_id_seq;
//CREATE TABLE public.in_app_message (
//	id bigint NOT NULL DEFAULT nextval('in_app_message_id_seq'),
//	group_ids jsonb NOT NULL,
//	type bigint  NOT NULL,
//	data jsonb NOT NULL,
//	start_date bigint,
//	end_date bigint,
//	high_priority bigint NOT NULL,
//	create_time timestamp with time zone NOT NULL DEFAULT now(),
//	update_time timestamp with time zone NOT NULL DEFAULT now(),
//	constraint in_app_message_pk primary key (id)
//);
//ALTER SEQUENCE in_app_message_id_seq OWNED BY public.in_app_message.id;
const InAppMessageTableName = "in_app_message"

func AddInAppMessage(ctx context.Context, logger runtime.Logger, db *sql.DB, marshaler *protojson.MarshalOptions, inAppMessage *pb.InAppMessage) (*pb.InAppMessage, error) {
	if inAppMessage == nil || inAppMessage.Type < 0 || inAppMessage.Data == nil {
		return nil, status.Error(codes.InvalidArgument, "Error add inAppMessage.")
	}
	data, _ := marshaler.Marshal(inAppMessage.Data)
	group := "["
	if len(inAppMessage.GroupIds) > 0 {
		for idx, groupId := range inAppMessage.GroupIds {
			if idx == 0 {
				group += fmt.Sprintf(`%d`, groupId)
			} else {
				group += fmt.Sprintf(`, %d`, groupId)
			}
		}
	}
	group += "]"
	query := "INSERT INTO " + InAppMessageTableName + " (group_ids, type, data, start_date, end_date, high_priority, create_time, update_time) VALUES ($1, $2, $3, $4, $5, $6, now(), now())"
	result, err := db.ExecContext(ctx, query, group, inAppMessage.Type, data,
		inAppMessage.StartDate,
		inAppMessage.EndDate,
		inAppMessage.HighPriority)
	if err != nil {
		logger.Error("Add inAppMessage, type: %v, groupId: %v, data: %v, error %s",
			inAppMessage.Type, inAppMessage.GroupIds, inAppMessage.Data, err.Error())
		return nil, status.Error(codes.Internal, "Error add inAppMessage.")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert inAppMessage, type: %v, groupId: %v, data: %v",
			inAppMessage.Type, inAppMessage.GroupIds, inAppMessage.Data)
		return nil, status.Error(codes.Internal, "Error add inAppMessage.")
	}
	return inAppMessage, nil
}

func GetInAppMessageById(ctx context.Context, logger runtime.Logger, db *sql.DB, unmarshaler *protojson.UnmarshalOptions, id int64) (*pb.InAppMessage, error) {
	if id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Id is empty")
	}
	query := "SELECT id, group_ids, type, data, start_date, end_date, high_priority FROM " + InAppMessageTableName + " WHERE id=$1"
	var dbID int64
	var dbType int32
	var groupIdsStr string
	var dbData []byte
	var dbStartDate, dbEndDate, dbHighPriority int64
	err := db.QueryRowContext(ctx, query, id).Scan(&dbID, &groupIdsStr, &dbType, &dbData, &dbStartDate, &dbEndDate, &dbHighPriority)
	if err != nil {
		logger.Error("Query inAppMessage by id %d, error %s", id, err.Error())
		return nil, status.Error(codes.Internal, "Query inAppMessage error")
	}
	var data = &pb.InAppMessageData{}
	err = unmarshaler.Unmarshal(dbData, data)
	if err != nil {
		logger.Error("Unmarshal inAppMessage error %s", id, err.Error())
		return nil, status.Error(codes.Internal, "Query inAppMessage error")
	}
	var groupIds []int64
	_ = json.Unmarshal([]byte(groupIdsStr), &groupIds)
	inAppMessage := pb.InAppMessage{
		Id:           dbID,
		GroupIds:     groupIds,
		Type:         pb.TypeInAppMessage(dbType),
		Data:         data,
		StartDate:    dbStartDate,
		EndDate:      dbEndDate,
		HighPriority: dbHighPriority,
	}
	return &inAppMessage, nil
}

func DeleteInAppMessage(ctx context.Context, logger runtime.Logger, db *sql.DB, id int64) error {
	query := "DELETE FROM " + InAppMessageTableName + " WHERE id=$1"
	result, err := db.ExecContext(ctx, query, id)
	if err != nil {
		logger.Error("Delete inAppMessage by id %d, error %s", id, err.Error())
		return status.Error(codes.Internal, "Delete inAppMessage error")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did delete user group")
		return status.Error(codes.Internal, "Error delete inAppMessage")
	}
	return nil
}

func UpdateInAppMessage(ctx context.Context, logger runtime.Logger, db *sql.DB, marshaler *protojson.MarshalOptions, unmarshaler *protojson.UnmarshalOptions, id int64, inAppMessage *pb.InAppMessage) (*pb.InAppMessage, error) {
	oldInAppMessage, err := GetInAppMessageById(ctx, logger, db, unmarshaler, id)
	if err != nil {
		return nil, err
	}
	data, _ := marshaler.Marshal(inAppMessage.Data)

	group := "["
	if len(inAppMessage.GroupIds) > 0 {
		for idx, groupId := range inAppMessage.GroupIds {
			if idx == 0 {
				group += fmt.Sprintf(`%d`, groupId)
			} else {
				group += fmt.Sprintf(`, %d`, groupId)
			}
		}
	}
	group += "]"
	query := "UPDATE " + InAppMessageTableName + " SET group_ids=$1, data=$2, start_date=$3, end_date=$4, high_priority=$5 WHERE id=$6"
	result, err := db.ExecContext(ctx, query, group, data, inAppMessage.StartDate, inAppMessage.EndDate, inAppMessage.HighPriority, oldInAppMessage.Id)
	if err != nil {
		logger.Error("Update inAppMessage id %d, error %s", id, err.Error())
		return nil, status.Error(codes.Internal, "Update inAppMessage error")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not update inAppMessage")
		return nil, status.Error(codes.Internal, "Error Update inAppMessage")
	}
	oldInAppMessage.GroupIds = inAppMessage.GroupIds
	oldInAppMessage.StartDate = inAppMessage.StartDate
	oldInAppMessage.EndDate = inAppMessage.EndDate
	oldInAppMessage.Data = inAppMessage.Data
	return oldInAppMessage, nil
}

func GetListInAppMessage(ctx context.Context, logger runtime.Logger, db *sql.DB, unmarshaler *protojson.UnmarshalOptions, nk runtime.NakamaModule, limit int64, cursor string, typeInAppMessage pb.TypeInAppMessage) (*pb.ListInAppMessage, error) {
	var incomingCursor = &entity.InAppMessageListCursor{}
	if cursor != "" {
		cb, err := base64.URLEncoding.DecodeString(cursor)
		if err != nil {
			return nil, ErrWalletLedgerInvalidCursor
		}
		if err := gob.NewDecoder(bytes.NewReader(cb)).Decode(incomingCursor); err != nil {
			return nil, ErrWalletLedgerInvalidCursor
		}
		logger.Info("GetListInAppMessage with cusor %d, create time %s ",
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
	params = append(params, typeInAppMessage)
	userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	logger.Debug("UserId %v", userID)

	if userID != "" {
		allGroups, _ := GetAllGroupByUser(ctx, logger, db, nk, userID)
		params = append(params, time.Now().Unix())
		params = append(params, time.Now().Unix())
		query += " WHERE type=$1 and start_date <= $2"
		count := 0
		if incomingCursor.Id > 0 {
			count = 5
			if incomingCursor.IsNext {
				query += " and (end_date = 0 or end_date >= $3) and id < $4 "
				if len(allGroups) > 0 {
					query += " and ARRAY(SELECT jsonb_array_elements(group_ids)::jsonb) <@ ARRAY(SELECT json_array_elements('["
					for idx, idGroup := range allGroups {
						if idx == 0 {
							query += fmt.Sprintf(" $%d ", idGroup)
						} else {
							query += fmt.Sprintf(", $%d ", idGroup)
						}
						//params = append(params, idGroup)
						//count++
					}
					query += " ]')::jsonb)"
				}
				query += "order by high_priority desc, id desc"
			} else {
				query += " and (end_date = 0 or end_date >= $3) and id > $4 "
				if len(allGroups) > 0 {
					query += " and ARRAY(SELECT jsonb_array_elements(group_ids)::jsonb) <@ ARRAY(SELECT json_array_elements('["
					for idx, idGroup := range allGroups {
						if idx == 0 {
							query += fmt.Sprintf(" $%d ", idGroup)
						} else {
							query += fmt.Sprintf(", $%d ", idGroup)
						}
						//params = append(params, idGroup)
						//count++
					}
					query += " ]')::jsonb)"
				}
				query += "order by high_priority desc, id asc"
			}
			params = append(params, incomingCursor.Id)
			query += fmt.Sprintf("  limit $%d", count)
			params = append(params, limit)
		} else {
			count = 4
			query += " and (end_date = 0 or end_date >= $3)"
			if len(allGroups) > 0 {
				query += " and ARRAY(SELECT jsonb_array_elements(group_ids)::jsonb) <@ ARRAY(SELECT json_array_elements('["
				for idx, idGroup := range allGroups {
					if idx == 0 {
						query += fmt.Sprintf(" %d ", idGroup)
					} else {
						query += fmt.Sprintf(", %d ", idGroup)
					}
					//params = append(params, idGroup)
					//count++
				}
				query += " ]')::jsonb)"
			}
			query += fmt.Sprintf(" order by high_priority desc, id desc limit $%d", count)
			params = append(params, limit)
		}
	} else {
		if incomingCursor.Id > 0 {
			if incomingCursor.IsNext {
				query += " WHERE type=$1 and id < $2 order by high_priority desc, id desc"
			} else {
				query += " WHERE type=$1 and id > $2 order by high_priority desc, id asc"
			}
			params = append(params, incomingCursor.Id)
			query += "  limit $3"
			params = append(params, limit)
		} else {
			query += " WHERE type=$1 order by high_priority desc, id desc limit $2"
			params = append(params, limit)
		}
	}

	queryRow := "SELECT id, group_ids, type, data, start_date, end_date, high_priority FROM " +
		InAppMessageTableName + query

	logger.Debug("queryRow %s %v", queryRow, params)
	rows, err = db.QueryContext(ctx, queryRow, params...)
	if err != nil {
		logger.Error("Query lists inAppMessage, error %s", err.Error())
		return nil, status.Error(codes.Internal, "Query lists inAppMessage")
	}
	ml := make([]*pb.InAppMessage, 0)
	var dbID int64
	var dbType int32
	var dbData []byte
	var groupIds []int64
	var dbStartDate, dbEndDate, dbHighPriority int64
	if err != nil {
		logger.Error("Query inAppMessage error %s", err.Error())
		return nil, status.Error(codes.Internal, "Query inAppMessage error")
	}
	for rows.Next() {
		var groupIdsStr string
		rows.Scan(&dbID, &groupIdsStr, &dbType, &dbData, &dbStartDate, &dbEndDate, &dbHighPriority)
		var data = &pb.InAppMessageData{}
		err = unmarshaler.Unmarshal(dbData, data)
		if err != nil {
			logger.Error("Unmarshal inAppMessage error %s", err.Error())
			return nil, status.Error(codes.Internal, "Query inAppMessage error")
		}
		_ = json.Unmarshal([]byte(groupIdsStr), &groupIds)
		inAppMessage := pb.InAppMessage{
			Id:           dbID,
			GroupIds:     groupIds,
			Type:         pb.TypeInAppMessage(dbType),
			Data:         data,
			StartDate:    dbStartDate,
			EndDate:      dbEndDate,
			HighPriority: dbHighPriority,
		}
		ml = append(ml, &inAppMessage)
	}
	var total int64 = incomingCursor.Total
	if total <= 0 {
		queryTotal := "Select count(*) as total FROM " + InAppMessageTableName +
			strings.ReplaceAll(query, "order by high_priority desc, id desc", "")

		_ = db.QueryRowContext(ctx, queryTotal, params...).Scan(&total)
	}
	var nextCursor *entity.InAppMessageListCursor
	var prevCursor *entity.InAppMessageListCursor
	if len(ml) > 0 {
		if len(ml)+int(incomingCursor.Offset) < int(total) {
			nextCursor = &entity.InAppMessageListCursor{
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
		prevCursor = &entity.InAppMessageListCursor{
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
			logger.Error("Error creating list cursor", zap.Error(err))
			return nil, err
		}
		nextCursorStr = base64.URLEncoding.EncodeToString(cursorBuf.Bytes())
	}

	var prevCursorStr string
	if prevCursor != nil {
		cursorBuf := new(bytes.Buffer)
		if err := gob.NewEncoder(cursorBuf).Encode(prevCursor); err != nil {
			logger.Error("Error creating list cursor", zap.Error(err))
			return nil, err
		}
		prevCursorStr = base64.URLEncoding.EncodeToString(cursorBuf.Bytes())
	}

	return &pb.ListInAppMessage{
		InAppMessages: ml,
		NextCusor:     nextCursorStr,
		PrevCusor:     prevCursorStr,
		Total:         total,
		Offset:        incomingCursor.Offset,
		Limit:         limit,
	}, nil
}
