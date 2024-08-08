package cgbdb

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"sort"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgtype"
	"github.com/nakamaFramework/cgb-lobby-module/entity"
	pb "github.com/nakamaFramework/cgp-common/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CREATE SEQUENCE cgb_notification_id_seq;
// CREATE TABLE public.cgb_notification (
//
//	id bigint NOT NULL DEFAULT nextval('notification_id_seq'),
//	title character varying(256)  NOT NULL,
//	content text NOT NULL,
//	sender_id character varying(128) NOT NULL,
//	recipient_id character varying(128) NOT NULL,
//	type bigint  NOT NULL,
//	read boolean NOT NULL,
//	create_time timestamp with time zone NOT NULL DEFAULT now(),
//	update_time timestamp with time zone NOT NULL DEFAULT now(),
//	constraint cgb_notification_pk primary key (id)
// );
// ALTER SEQUENCE cgb_notification_id_seq OWNED BY public.cgb_notification.id;
// ALTER TABLE public.cgb_notification ADD COLUMN app_package text NULL;
// ALTER TABLE public.cgb_notification ADD COLUMN game_id text NULL;

const NotificationTableName = "cgb_notification"

func AddNotification(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, notification *pb.Notification) error {
	if notification == nil || notification.Title == "" || notification.Type <= 0 || notification.Content == "" || notification.RecipientId == "" {
		return status.Error(codes.InvalidArgument, "Error add notification.")
	}
	query := "INSERT INTO " + NotificationTableName + " (title, content, sender_id, recipient_id, type, read, app_package, game_id, create_time, update_time) VALUES ($1, $2, $3, $4, $5, false, $6, $7, now(), now())"
	result, err := db.ExecContext(ctx, query, notification.Title, notification.Content, notification.SenderId, notification.RecipientId, notification.Type, notification.AppPackage, notification.GameId)
	if err != nil {
		logger.Error("Add notification, type: %s, title: %s, content: %s, error %s",
			notification.Type, notification.Title, notification.Content, err.Error())
		return status.Error(codes.Internal, "Error add notification.")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert new notification, type: %s, title: %s, content: %s",
			notification.Type, notification.Title, notification.Content)
		return status.Error(codes.Internal, "Error add notification.")
	}
	content := map[string]interface{}{
		"content": notification.Content,
	}
	err = nk.NotificationSend(ctx, notification.RecipientId, notification.Title, content, (int)(notification.Type), notification.SenderId, false)
	if err != nil {
		logger.Error("Send notification type: %s, title: %s, content: %s, error %s",
			notification.Type, notification.Title, notification.Content, err.Error())
	}
	return err
}

func GetNotificationById(ctx context.Context, logger runtime.Logger, db *sql.DB, id int64, user_id string) (*pb.Notification, error) {
	if id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Id is empty")
	}
	query := "SELECT id, title, content, sender_id, recipient_id, type, read, create_time FROM " + NotificationTableName + " WHERE id=$1 and recipient_id=$2"
	var dbID int64
	var dbTitle, dbContent, dbSenderId, dbRecipientId string
	var dbType int32
	var dbRead bool
	var dbCreateTime pgtype.Timestamptz

	err := db.QueryRowContext(ctx, query, id, user_id).
		Scan(&dbID, &dbTitle, &dbContent,
			&dbSenderId, &dbRecipientId, &dbType,
			&dbRead, &dbCreateTime)
	if err != nil {
		logger.Error("Query notification by id %, error %s", id, err.Error())
		return nil, status.Error(codes.Internal, "Query notification error")
	}
	notification := pb.Notification{
		Id:             dbID,
		Title:          dbTitle,
		Content:        dbContent,
		SenderId:       dbSenderId,
		RecipientId:    dbRecipientId,
		Type:           (pb.TypeNotification)(dbType),
		Read:           dbRead,
		CreateTimeUnix: dbCreateTime.Time.Unix(),
	}
	return &notification, nil
}

func ReadNotification(ctx context.Context, logger runtime.Logger, db *sql.DB, id int64, user_id string) error {
	if !IsExistNotificationNotRead(ctx, logger, db, user_id) {
		return nil
	}
	query := "UPDATE " + NotificationTableName + " SET read=$1 WHERE id=$2 and recipient_id=$3"
	result, err := db.ExecContext(ctx, query, true, id, user_id)
	if err != nil {
		logger.Error("Update user group id %d, user %s, error %s", id, err.Error())
		return status.Error(codes.Internal, "Read notification error")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount == 0 {
		logger.Error("Did not update notification")
		return status.Error(codes.Internal, "Read notification group")
	}
	return nil
}

func DeleteNotification(ctx context.Context, logger runtime.Logger, db *sql.DB, id int64, userId string) error {
	query := "DELETE FROM " + NotificationTableName + " WHERE id=$1 and recipient_id=$2"
	result, err := db.ExecContext(ctx, query, id, userId)
	if err != nil {
		logger.Error("Delete notification by id %d, error %s", id, err.Error())
		return status.Error(codes.Internal, "Delete notification error")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did delete notification")
		return status.Error(codes.Internal, "Error delete notification")
	}
	return nil
}

func ReadAllNotification(ctx context.Context, logger runtime.Logger, db *sql.DB, user_id string) error {
	if !IsExistNotificationNotRead(ctx, logger, db, user_id) {
		return nil
	}
	query := "UPDATE " + NotificationTableName + " SET read=$1 WHERE recipient_id=$2"
	result, err := db.ExecContext(ctx, query, true, user_id)
	if err != nil {
		logger.Error("Read all notification, user %s, error %s", user_id, err.Error())
		return status.Error(codes.Internal, "Read all notification error")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount == 0 {
		logger.Error("Did not update notification")
		return status.Error(codes.Internal, "Read notification group")
	}
	return nil
}

func DeleteAllNotification(ctx context.Context, logger runtime.Logger, db *sql.DB, userId string) error {
	query := "DELETE FROM " + NotificationTableName + " WHERE recipient_id=$1"
	result, err := db.ExecContext(ctx, query, userId)
	if err != nil {
		logger.Error("Delete all notification, error %s", err.Error())
		return status.Error(codes.Internal, "Delete notification error")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did delete notification")
		return status.Error(codes.Internal, "Error delete notification")
	}
	return nil
}

func GetListNotification(ctx context.Context, logger runtime.Logger, db *sql.DB, limit int64, cursor string, userId string, typeNotification pb.TypeNotification) (*pb.ListNotification, error) {
	var incomingCursor = &entity.NotificationListCursor{}
	if cursor != "" {
		cb, err := base64.URLEncoding.DecodeString(cursor)
		if err != nil {
			return nil, ErrWalletLedgerInvalidCursor
		}
		if err := gob.NewDecoder(bytes.NewReader(cb)).Decode(incomingCursor); err != nil {
			return nil, ErrWalletLedgerInvalidCursor
		}
		if userId != incomingCursor.UserId {
			return nil, ErrWalletLedgerInvalidCursor
		}
		logger.Info("GetListNotification with cusor %d, create time %s ",
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
	params = append(params, userId)
	params = append(params, typeNotification)

	if incomingCursor.Id > 0 {
		if incomingCursor.IsNext {
			query += " WHERE recipient_id=$1 and type=$2 and id < $3 order by id desc "
		} else {
			query += " WHERE recipient_id=$1 and type=$2 and id > $3 AND deleted = false order by id asc"
		}
		params = append(params, incomingCursor.Id)
		query += "  limit $4"
		params = append(params, limit)
	} else {
		query += " WHERE recipient_id=$1 and type=$2 order by id desc limit $3"
		params = append(params, limit)
	}
	queryRow := "SELECT id, title, content, sender_id, recipient_id, type, read, app_package, game_id, create_time FROM " +
		NotificationTableName + query
	rows, err = db.QueryContext(ctx, queryRow, params...)
	if err != nil {
		logger.Error("Query lists notification, error %s", err.Error())
		return nil, status.Error(codes.Internal, "Query lists notification")
	}
	ml := make([]*pb.Notification, 0)
	var dbID int64
	var dbTitle, dbContent, dbSenderId, dbRecipientId, dbAppPackage, dbGameId string
	var dbType int32
	var dbRead bool
	var dbCreateTime pgtype.Timestamptz

	for rows.Next() {
		rows.Scan(&dbID, &dbTitle, &dbContent,
			&dbSenderId, &dbRecipientId,
			&dbType, &dbRead, &dbAppPackage, &dbGameId, &dbCreateTime)
		notification := pb.Notification{
			Id:             dbID,
			RecipientId:    dbRecipientId,
			Type:           (pb.TypeNotification)(dbType),
			Title:          dbTitle,
			Content:        dbContent,
			SenderId:       dbSenderId,
			Read:           dbRead,
			CreateTimeUnix: dbCreateTime.Time.Unix(),
			AppPackage:     dbAppPackage,
			GameId:         dbGameId,
		}
		ml = append(ml, &notification)
	}
	sort.Slice(ml, func(i, j int) bool {
		return ml[i].Id > ml[j].Id
	})
	var total int64 = incomingCursor.Total
	if total <= 0 {
		queryTotal := "Select count(*) as total FROM " + NotificationTableName +
			strings.ReplaceAll(query, "order by id desc", "")

		_ = db.QueryRowContext(ctx, queryTotal, params...).Scan(&total)
	}
	var nextCursor *entity.NotificationListCursor
	var prevCursor *entity.NotificationListCursor
	if len(ml) > 0 {
		if len(ml)+int(incomingCursor.Offset) < int(total) {
			nextCursor = &entity.NotificationListCursor{
				Id:     ml[len(ml)-1].Id,
				IsNext: true,
				UserId: userId,
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
		prevCursor = &entity.NotificationListCursor{
			Id:     ml[0].Id,
			UserId: userId,
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

	return &pb.ListNotification{
		Notifications: ml,
		NextCusor:     nextCursorStr,
		PrevCusor:     prevCursorStr,
		Total:         total,
		Offset:        incomingCursor.Offset,
		Limit:         limit,
	}, nil
}

func IsExistNotificationNotRead(ctx context.Context, logger runtime.Logger, db *sql.DB, user_id string) bool {
	query := "SELECT id FROM " + NotificationTableName + " WHERE recipient_id=$1 and read=false LIMIT 1"
	var dbID int64

	err := db.QueryRowContext(ctx, query, user_id).Scan(&dbID)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		logger.Error("Query IsExistNotificationNotRead user id %s, error %s", user_id, err.Error())
		return false
	}
	return true
}
