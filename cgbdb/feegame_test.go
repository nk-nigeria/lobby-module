package cgbdb

import (
	"context"
	"database/sql"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
	_ "github.com/lib/pq"
	"github.com/nakamaFramework/cgb-lobby-module/entity"
)

func TestAddNewFeeGame(t *testing.T) {
	type args struct {
		ctx     context.Context
		logger  runtime.Logger
		db      *sql.DB
		feeGame entity.FeeGame
	}
	connStr := "postgresql://postgres:localdb@127.0.0.1/nakama?sslmode=disable"
	mdb, _ := sql.Open("postgres", connStr)
	// userId := "b06fb31e-6fba-44ae-b08f-01286eaf9b79"
	// userId := "115343e1-5e87-4774-976a-f5c6923b80be"
	userId := "548b7190-b8e3-47d6-8b2f-6321c35cacb2"
	beginW, _ := entity.RangeThisWeek()
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test-add-fee-game-1",
			args: args{
				ctx:    context.Background(),
				logger: &entity.EmptyLogger{},
				db:     mdb,
				feeGame: entity.FeeGame{
					UserID:         userId,
					Fee:            int64(entity.RandomInt(1000, 10000)),
					Game:           "",
					CreateTimeUnix: beginW.Unix() + int64(entity.RandomInt(0, 86400*6)),
				},
			},
		},
		{
			name: "test-add-fee-game-2",
			args: args{
				ctx:    context.Background(),
				logger: &entity.EmptyLogger{},
				db:     mdb,
				feeGame: entity.FeeGame{
					UserID:         userId,
					Fee:            int64(entity.RandomInt(1000, 10000)),
					Game:           "",
					CreateTimeUnix: beginW.Unix() + int64(entity.RandomInt(0, 86400*6)),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddNewFeeGame(tt.args.ctx, tt.args.logger, tt.args.db, tt.args.feeGame)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddNewFeeGame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AddNewFeeGame() = %v, want %v", got, tt.want)
			}
		})
	}
}
