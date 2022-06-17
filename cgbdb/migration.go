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

	_, err = db.ExecContext(ctx, `CREATE TABLE public.freechip (
		id bigint NOT NULL,
		sender_id character varying(128) NOT NULL,
		recipient_id character varying(128) NOT NULL,
		title character varying(128) NOT NULL,
		content character varying(128) NOT NULL,
		chips integer NOT NULL DEFAULT 0,
		claimable smallint NOT NULL DEFAULT 1,
		create_time timestamp with time zone NOT NULL DEFAULT now(),
		update_time timestamp with time zone NOT NULL DEFAULT now()
		);
		ALTER TABLE
		public.freechip
		ADD
		CONSTRAINT freechip_pkey PRIMARY KEY (id)
	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())
		return
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE
		public.giftcode (
			id bigint NOT NULL,
			code character varying(128) NOT NULL DEFAULT '',
			UNIQUE(code),
			n_current integer NOT NULL DEFAULT 0,
			n_max integer NOT NULL DEFAULT 0,
			value integer NOT NULL DEFAULT 0,
			start_time_unix timestamp,
			end_time_unix timestamp,
			message character varying(256) NOT NULL DEFAULT '',
			vip integer NOT NULL DEFAULT 0,
			gift_code_type smallint NOT NULL DEFAULT 1,
			deleted smallint NOT NULL DEFAULT 0,
			create_time timestamp

			with
			time zone NOT NULL DEFAULT now(),
			update_time timestamp
			with
			time zone NOT NULL DEFAULT now()
		);

		ALTER TABLE
		public.giftcode
		ADD
		CONSTRAINT giftcode_pkey PRIMARY KEY (id)
	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())
		return
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE
		public.giftcodeclaim (
			id bigint NOT NULL,
			code character varying(128) NOT NULL DEFAULT '',
			user_id character varying(128) NOT NULL,
			create_time timestamp
			with
			time zone NOT NULL DEFAULT now(),
			update_time timestamp
			with
			time zone NOT NULL DEFAULT now()
		);

		ALTER TABLE
		public.giftcodeclaim
		ADD
  		CONSTRAINT giftcodeclaim_pkey PRIMARY KEY (id)
	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())
		return
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE public.exchange (
		id bigint NOT NULL,
		id_deal character varying(128) NOT NULL,
			chips integer NOT NULL DEFAULT 0,
		price character varying(128) NOT NULL,
		status smallint NOT NULL DEFAULT 0,
		unlock smallint NOT NULL DEFAULT 1,
		cash_id character varying(128) NOT NULL,
		cash_type character varying(128) NOT NULL,
		user_id_request character varying(128) NOT NULL,
		user_name_request character varying(128) NOT NULL,
		vip_lv smallint NOT NULL DEFAULT 0,
		device_id character varying(128) NOT NULL,
		user_id_handling character varying(128) NOT NULL,
		user_name_handling character varying(128) NOT NULL,
		reason character varying(128) NOT NULL,
		create_time timestamp with time zone NOT NULL DEFAULT now(),
		update_time timestamp with time zone NOT NULL DEFAULT now()
		);
		ALTER TABLE
		public.exchange
		ADD
		CONSTRAINT exchange_pkey PRIMARY KEY (id)
	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())
		return
	}

	logger.Error("Done run migration")
}
