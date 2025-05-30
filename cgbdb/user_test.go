package cgbdb

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAccount(t *testing.T) {
	name := "GetAccount"
	connStr := "user=postgres dbname=nakama password=localdb host=103.226.250.195 port=6432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	assert.NoError(t, err)
	defer db.Close()
	t.Run(name, func(t *testing.T) {
		{
			account, err := GetAccount(context.Background(), db, "701494a4-c6d4-4540-b377-ebb1981fbeb1", 0)
			assert.NoError(t, err)
			assert.NotNil(t, account)
		}
		{
			account, err := GetAccount(context.Background(), db, "", 2000048)
			assert.NoError(t, err)
			assert.NotNil(t, account)
		}
		{
			account, err := GetAccount(context.Background(), db, "1", 0)
			assert.Error(t, err)
			assert.Nil(t, account)
		}
		{
			account, err := GetAccount(context.Background(), db, "", 1)
			assert.Error(t, err)
			assert.Nil(t, account)
		}
	})

}
