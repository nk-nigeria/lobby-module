package cgbdb

import (
	"context"
	"database/sql"

	"github.com/ciaolink-game-platform/cgp-lobby-module/conf"
	pb "github.com/ciaolink-game-platform/cgp-lobby-module/proto"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const FreeChipTableName = "freechip"

func AddClaimableFreeChip(ctx context.Context, logger runtime.Logger, db *sql.DB, freeChip *pb.FreeChip) error {
	if freeChip == nil || freeChip.RecipientId == "" || freeChip.Chips == 0 {
		return status.Error(codes.InvalidArgument, "Error add claimable freechip.")
	}
	freeChip.Id = conf.SnowlakeNode.Generate().String()
	query := "INSERT INTO " + FreeChipTableName + " (id, sender_id, recipient_id, title, content, chips, claimable, create_time, update_time) VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())"
	result, err := db.ExecContext(ctx, query, freeChip.Id, freeChip.SenderId, freeChip.RecipientId, freeChip.Title, freeChip.Content,
		freeChip.Chips, 1)
	if err != nil {
		logger.Error("Add new claimable, sender: %s, recv: %s, chips: %d, error %s",
			freeChip.SenderId, freeChip.RecipientId, freeChip.Chips, err.Error())
		return status.Error(codes.Internal, "Error add claimable freechip.")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not insert new claimable, sender: %s, recv: %s, chips: %s",
			freeChip.SenderId, freeChip.RecipientId, freeChip.Chips)
		return status.Error(codes.Internal, "Error add claimable freechip.")
	}
	return nil
}

func ClaimFreeChip(ctx context.Context, logger runtime.Logger, db *sql.DB, id string, userId string) (*pb.FreeChip, error) {
	freeChip, err := GetFreeChipByIdByUser(ctx, logger, db, id, userId)
	if err != nil {
		return nil, err
	}
	if freeChip.Claimable == false {
		return nil, status.Error(codes.Aborted, "Freechip alread claimed")
	}
	query := "UPDATE " + FreeChipTableName + " SET claimable=$1 WHERE id=$2 AND recipient_id=$3 AND claimable=$4"
	result, err := db.ExecContext(ctx, query, 0, id, userId, 1)
	if err != nil {
		logger.Error("Claim free chip id %d, user %s, error %s", id, userId, err.Error())
		return nil, status.Error(codes.Internal, "Claim freechip error")
	}
	if rowsAffectedCount, _ := result.RowsAffected(); rowsAffectedCount != 1 {
		logger.Error("Did not claim freechip.")
		return nil, status.Error(codes.Internal, "Error claim freechip")
	}
	return freeChip, nil
}

func GetFreeChipByIdByUser(ctx context.Context, logger runtime.Logger, db *sql.DB, id string, userId string) (*pb.FreeChip, error) {
	if id == "" || userId == "" {
		return nil, status.Error(codes.InvalidArgument, "Id or user id is empty")
	}
	query := "SELECT id, sender_id, recipient_id, title, content, chips, claimable FROM " + FreeChipTableName + " WHERE id=$1 AND recipient_id=$2"
	var dbID, dbSenderId, dbRecvId, dbTitle, dbContent string
	var dbChips int64
	var dbClaimable int
	err := db.QueryRowContext(ctx, query, id, userId).Scan(&dbID, &dbSenderId, &dbRecvId, &dbTitle, &dbContent, &dbChips, &dbClaimable)
	if err != nil {
		logger.Error("Query free chip id %, user %s, error %s", id, userId, err.Error())
		return nil, status.Error(codes.Internal, "Query freechip error")
	}
	freeChip := pb.FreeChip{
		Id:          dbID,
		SenderId:    dbSenderId,
		RecipientId: dbRecvId,
		Title:       dbTitle,
		Content:     dbContent,
		Chips:       dbChips,
	}
	if dbClaimable == 1 {
		freeChip.Claimable = true
	}
	return &freeChip, nil
}

func GetFreeChipClaimableByUser(ctx context.Context, logger runtime.Logger, db *sql.DB, userId string) (*pb.ListFreeChip, error) {
	if userId == "" {
		return nil, status.Error(codes.InvalidArgument, "Id or user id is empty")
	}
	query := "SELECT id, sender_id, recipient_id, title, content, chips, claimable FROM " + FreeChipTableName + " WHERE claimable=$1 AND recipient_id=$2"

	rows, err := db.QueryContext(ctx, query, 1, userId)
	if err != nil {
		logger.Error("Query free chip claimable user %s, error %s", userId, err.Error())
		return nil, status.Error(codes.Internal, "Query freechip claimable error")
	}
	ml := make([]*pb.FreeChip, 0)
	var dbID, dbSenderId, dbRecvId, dbTitle, dbContent string
	var dbChips int64
	var dbClaimable int
	for rows.Next() {
		rows.Scan(&dbID, &dbSenderId, &dbRecvId, &dbTitle, &dbContent, &dbChips, &dbClaimable)
		freeChip := pb.FreeChip{
			Id:          dbID,
			SenderId:    dbSenderId,
			RecipientId: dbRecvId,
			Title:       dbTitle,
			Content:     dbContent,
			Chips:       dbChips,
		}
		if dbClaimable == 1 {
			freeChip.Claimable = true
		}
		ml = append(ml, &freeChip)
	}

	return &pb.ListFreeChip{
		Lists: ml,
	}, nil
}
