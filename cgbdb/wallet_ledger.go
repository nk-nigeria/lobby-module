package cgbdb

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ciaolink-game-platform/cgp-lobby-module/entity"
	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgtype"
	"go.uber.org/zap"
)

var (
	ErrWalletLedgerInvalidCursor = errors.New("wallet ledger cursor invalid")
)

func ListWalletLedger(ctx context.Context, logger runtime.Logger, db *sql.DB, userID uuid.UUID, metaAction, metaBankAction []string, limit *int, cursor string) ([]runtime.WalletLedgerItem, string, string, error) {
	var incomingCursor *entity.WalletLedgerListCursor
	if cursor != "" {
		cb, err := base64.URLEncoding.DecodeString(cursor)
		if err != nil {
			return nil, "", "", ErrWalletLedgerInvalidCursor
		}
		incomingCursor = &entity.WalletLedgerListCursor{}
		if err := gob.NewDecoder(bytes.NewReader(cb)).Decode(incomingCursor); err != nil {
			return nil, "", "", ErrWalletLedgerInvalidCursor
		}

		// Cursor and filter mismatch. Perhaps the caller has sent an old cursor with a changed filter.
		if userID.String() != incomingCursor.UserId {
			return nil, "", "", ErrWalletLedgerInvalidCursor
		}
		logger.Info("ListWalletLedger with cusor, userId %s, Id %s, create time %s ", incomingCursor.UserId,
			incomingCursor.Id, incomingCursor.CreateTime.String())
	}
	params := []interface{}{userID, time.Now().UTC(), uuid.UUID{}}

	if incomingCursor != nil {
		params[1] = incomingCursor.CreateTime
		params[2], _ = uuid.FromString(incomingCursor.Id)
		metaAction = incomingCursor.MetaAction
		metaBankAction = incomingCursor.MetaBankAction
	}
	inMetaActionParam := ""
	for _, action := range metaAction {
		inMetaActionParam += fmt.Sprintf(`'%s',`, action)
	}
	if len(inMetaActionParam) > 0 {
		inMetaActionParam = inMetaActionParam[:len(inMetaActionParam)-1]
	}

	inMetaBankActionParam := ""
	for _, action := range metaBankAction {
		inMetaBankActionParam += fmt.Sprintf(`'%s',`, action)
	}
	if len(inMetaBankActionParam) > 0 {
		inMetaBankActionParam = inMetaBankActionParam[:len(inMetaBankActionParam)-1]
	}
	// params = append(params, inParam)

	query := `SELECT id, changeset, metadata, create_time, update_time 
	FROM wallet_ledger 
	WHERE user_id = $1::UUID 
	AND (user_id, create_time, id) < ($1::UUID, $2, $3::UUID) 
	AND metadata ->> 'action' IN  ( ` + inMetaActionParam + " ) "
	if len(inMetaBankActionParam) > 0 {
		query += `AND metadata ->> 'bank_action' IN  ( ` + inMetaBankActionParam + " ) "
	}
	query += ` ORDER BY create_time DESC`
	if incomingCursor != nil && !incomingCursor.IsNext {
		query = `SELECT id, changeset, metadata, create_time, update_time 
		FROM wallet_ledger 
		WHERE user_id = $1::UUID
		AND (user_id, create_time, id) > ($1::UUID, $2, $3::UUID) 
		AND metadata ->> 'action' IN  ( ` + inMetaActionParam + " ) "
		if len(inMetaBankActionParam) > 0 {
			query += `AND metadata ->> 'bank_action' IN  ( ` + inMetaBankActionParam + " ) "
		}
		query += ` ORDER BY create_time ASC`
	}

	if limit != nil {
		query = fmt.Sprintf(`%s LIMIT %v`, query, *limit+1)
	}

	logger.Info("Query ledger %s  params %v ", query, params)

	results := make([]runtime.WalletLedgerItem, 0, 10)
	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		logger.Error("Error retrieving user wallet ledger.", zap.String("user_id", userID.String()), zap.Error(err))
		return nil, "", "", err
	}
	defer rows.Close()

	var id string
	var changeset sql.NullString
	var metadata sql.NullString
	var createTime pgtype.Timestamptz
	var updateTime pgtype.Timestamptz
	var nextCursor *entity.WalletLedgerListCursor
	var prevCursor *entity.WalletLedgerListCursor
	for rows.Next() {
		if limit != nil && len(results) >= *limit {
			nextCursor = &entity.WalletLedgerListCursor{
				UserId:         userID.String(),
				Id:             id,
				CreateTime:     createTime.Time,
				IsNext:         true,
				MetaAction:     metaAction,
				MetaBankAction: metaBankAction,
			}
			break
		}

		err = rows.Scan(&id, &changeset, &metadata, &createTime, &updateTime)
		if err != nil {
			logger.Error("Error converting user wallet ledger.", zap.String("user_id", userID.String()), zap.Error(err))
			return nil, "", "", err
		}

		var changesetMap map[string]int64
		err = json.Unmarshal([]byte(changeset.String), &changesetMap)
		if err != nil {
			logger.Error("Error converting user wallet ledger changeset.", zap.String("user_id", userID.String()), zap.Error(err))
			return nil, "", "", err
		}

		var metadataMap map[string]interface{}
		err = json.Unmarshal([]byte(metadata.String), &metadataMap)
		if err != nil {
			logger.Error("Error converting user wallet ledger metadata.", zap.String("user_id", userID.String()), zap.Error(err))
			return nil, "", "", err
		}

		results = append(results, &entity.WalletLedger{
			ID:         id,
			Changeset:  changesetMap,
			Metadata:   metadataMap,
			CreateTime: createTime.Time.Unix(),
			UpdateTime: updateTime.Time.Unix(),
		})

		if incomingCursor != nil && prevCursor == nil {
			prevCursor = &entity.WalletLedgerListCursor{
				UserId:         userID.String(),
				Id:             id,
				CreateTime:     createTime.Time,
				IsNext:         false,
				MetaAction:     metaAction,
				MetaBankAction: metaBankAction,
			}
		}
	}

	if incomingCursor != nil && !incomingCursor.IsNext {
		if nextCursor != nil && prevCursor != nil {
			nextCursor, nextCursor.IsNext, prevCursor, prevCursor.IsNext = prevCursor, prevCursor.IsNext, nextCursor, nextCursor.IsNext
		} else if nextCursor != nil {
			nextCursor, prevCursor = nil, nextCursor
			prevCursor.IsNext = !prevCursor.IsNext
		} else if prevCursor != nil {
			nextCursor, prevCursor = prevCursor, nil
			nextCursor.IsNext = !nextCursor.IsNext
		}

		for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
			results[i], results[j] = results[j], results[i]
		}
	}

	var nextCursorStr string
	if nextCursor != nil {
		cursorBuf := new(bytes.Buffer)
		if err := gob.NewEncoder(cursorBuf).Encode(nextCursor); err != nil {
			logger.Error("Error creating wallet ledger list cursor", zap.Error(err))
			return nil, "", "", err
		}
		nextCursorStr = base64.URLEncoding.EncodeToString(cursorBuf.Bytes())
	}

	var prevCursorStr string
	if prevCursor != nil {
		cursorBuf := new(bytes.Buffer)
		if err := gob.NewEncoder(cursorBuf).Encode(prevCursor); err != nil {
			logger.Error("Error creating wallet ledger list cursor", zap.Error(err))
			return nil, "", "", err
		}
		prevCursorStr = base64.URLEncoding.EncodeToString(cursorBuf.Bytes())
	}

	return results, nextCursorStr, prevCursorStr, nil
}
