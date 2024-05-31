package cgbdb

import (
	"context"
	"database/sql"
)

func CreateNewUserbot(ctx context.Context, db *sql.DB, userId, gameCode string) error {
	stmt, err := db.Prepare("INSERT INTO public.users_bot (user_id, game_code) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(userId, gameCode)
	return err
}
