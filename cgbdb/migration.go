package cgbdb

import (
	"context"
	"database/sql"
	"github.com/heroiclabs/nakama-common/runtime"
)

func RunMigrations(ctx context.Context, logger runtime.Logger, db *sql.DB) {
	_, err := db.ExecContext(ctx, `
		CREATE SEQUENCE IF NOT EXISTS user_group_id_seq;

		CREATE TABLE IF NOT EXISTS public.user_group (
		  id bigint NOT NULL DEFAULT nextval('user_group_id_seq'),
		  name character varying(256) NOT NULL,
		  type character varying(128) NOT NULL,
		  data character varying(128) NOT NULL,
		  deleted boolean NOT NULL,
		  create_time timestamp with time zone NOT NULL DEFAULT now(),
		  update_time timestamp with time zone NOT NULL DEFAULT now(),
		  constraint user_group_pk primary key (id),
		  UNIQUE (name)
		);

		ALTER SEQUENCE user_group_id_seq OWNED BY public.user_group.id;
  `)
	if err != nil {
		logger.Error("Error: %s", err.Error())
		return
	}
	logger.Error("Done run migration")
}
