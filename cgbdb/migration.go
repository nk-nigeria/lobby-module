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

	_, err = db.ExecContext(ctx, `
		CREATE SEQUENCE IF NOT EXISTS cgb_notification_id_seq;
		CREATE TABLE IF NOT EXISTS public.cgb_notification (
			id bigint NOT NULL DEFAULT nextval('cgb_notification_id_seq'),
			title character varying(256)  NOT NULL,
			content text NOT NULL,
			sender_id character varying(128) NOT NULL,
			recipient_id character varying(128) NOT NULL,
			type bigint  NOT NULL,
			read boolean NOT NULL,
			create_time timestamp with time zone NOT NULL DEFAULT now(),
			update_time timestamp with time zone NOT NULL DEFAULT now(),
			constraint cgb_notification_pk primary key (id)
		);
		ALTER SEQUENCE cgb_notification_id_seq OWNED BY public.cgb_notification.id;
  	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())
		return
	}

	_, err = db.ExecContext(ctx, `
		CREATE SEQUENCE IF NOT EXISTS in_app_message_id_seq;
		CREATE TABLE IF NOT EXISTS public.in_app_message (
			id bigint NOT NULL DEFAULT nextval('in_app_message_id_seq'),
			group_id bigint NOT NULL,
			type bigint  NOT NULL,
			data jsonb NOT NULL,
			start_date bigint,
			end_date bigint,
			high_priority bigint NOT NULL,
			create_time timestamp with time zone NOT NULL DEFAULT now(),
			update_time timestamp with time zone NOT NULL DEFAULT now(),
			constraint in_app_message_pk primary key (id)
		);
		ALTER SEQUENCE in_app_message_id_seq OWNED BY public.in_app_message.id;
  	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())
		return
	}
	logger.Error("Done run migration")
}
