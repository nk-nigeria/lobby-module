package cgbdb

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
)

func RunMigrations(ctx context.Context, logger runtime.Logger, db *sql.DB) {

	_, err := db.ExecContext(ctx, `
		CREATE SEQUENCE IF NOT EXISTS users_ext_sid_seq;

		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users_ext') THEN
				CREATE TABLE public.users_ext (
					id UUID PRIMARY KEY REFERENCES public.users(id) ON DELETE CASCADE,
					sid BIGINT DEFAULT nextval('users_ext_sid_seq')
				);
				ALTER SEQUENCE users_ext_sid_seq OWNED BY public.users_ext.sid;
			END IF;
		END$$;
	`)
	if err != nil {
		logger.Error("Error creating users_ext table: %s", err.Error())
	}

	_, err = db.ExecContext(ctx, `
		CREATE SEQUENCE IF NOT EXISTS user_group_id_seq;

		CREATE TABLE IF NOT EXISTS public.user_group (
		  id bigint NOT NULL DEFAULT nextval('user_group_id_seq'),
		  name character varying(256) NOT NULL,
		  create_time timestamp with time zone NOT NULL DEFAULT now(),
		  update_time timestamp with time zone NOT NULL DEFAULT now(),
		  constraint user_group_pk primary key (id),
		  UNIQUE (name)
		);

		ALTER SEQUENCE user_group_id_seq OWNED BY public.user_group.id;
	
		
  `)
	if err != nil {
		logger.Error("Error: %s", err.Error())

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
			app_package text NULL,
	        game_id text NULL,
			constraint cgb_notification_pk primary key (id)
		);
		ALTER SEQUENCE cgb_notification_id_seq OWNED BY public.cgb_notification.id;
  	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())

	}

	_, err = db.ExecContext(ctx, `
		CREATE SEQUENCE IF NOT EXISTS in_app_message_id_seq;
		CREATE TABLE IF NOT EXISTS public.in_app_message (
			id bigint NOT NULL DEFAULT nextval('in_app_message_id_seq'),
			group_ids jsonb NOT NULL,
			type bigint  NOT NULL,
			data jsonb NOT NULL,
			start_date bigint,
			end_date bigint,
			high_priority bigint NOT NULL,
			create_time timestamp with time zone NOT NULL DEFAULT now(),
			update_time timestamp with time zone NOT NULL DEFAULT now(),
			app_package text NULL,
	        game_id text NULL,
			constraint in_app_message_pk primary key (id)
		);
		ALTER SEQUENCE in_app_message_id_seq OWNED BY public.in_app_message.id;
  	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())

	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS public.freechip (
		id BIGINT NOT NULL PRIMARY KEY,
		sender_id character varying(128) NOT NULL,
		recipient_id character varying(128) NOT NULL,
		title character varying(128) NOT NULL,
		content character varying(128) NOT NULL,
		chips integer NOT NULL DEFAULT 0,
		claimable smallint NOT NULL DEFAULT 1,
		action character varying(128) NOT NULL,
		create_time timestamp with time zone NOT NULL DEFAULT now(),
		update_time timestamp with time zone NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())

	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS
		public.giftcode (
			id bigint PRIMARY KEY,
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
			create_time timestamp
			with
			time zone NOT NULL DEFAULT now(),
			update_time timestamp
			with
			time zone NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())

	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS
		public.giftcodeclaim (
			id bigint PRIMARY KEY,
			id_code bigint NOT NULL,
			code character varying(128) NOT NULL DEFAULT '',
			user_id character varying(128) NOT NULL,
			create_time timestamp
			with
			time zone NOT NULL DEFAULT now(),
			update_time timestamp
			with
			time zone NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())

	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS public.exchange (
		id bigint PRIMARY KEY,
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
		)
	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS
		public.giftcodetombstone (
			id bigint NOT NULL,
			code character varying(128) NOT NULL DEFAULT '',
			n_current integer NOT NULL DEFAULT 0,
			n_max integer NOT NULL DEFAULT 0,
			value integer NOT NULL DEFAULT 0,
			start_time_unix timestamp,
			end_time_unix timestamp,
			message character varying(256) NOT NULL DEFAULT '',
			vip integer NOT NULL DEFAULT 0,
			gift_code_type smallint NOT NULL DEFAULT 1,
			create_time timestamp
			with
			time zone NOT NULL DEFAULT now(),
			update_time timestamp
			with
			time zone NOT NULL DEFAULT now(),
			CONSTRAINT giftcodetombstone_pkey PRIMARY KEY (id)
		)
	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())

	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS
		public.referuser (
			id bigint NOT NULL,
			user_invitor character varying(128) NOT NULL,
			user_invitee character varying(128) NOT NULL,
				UNIQUE(user_invitee),
			create_time timestamp
			with
			time zone NOT NULL DEFAULT now(),
			update_time timestamp
			with
			time zone NOT NULL DEFAULT now(),
			CONSTRAINT referuser_pkey PRIMARY KEY (id)
		)
	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())

	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS
		public.reward_refer (
			id bigint NOT NULL,
			user_id character varying(128) NOT NULL,
			win_amt bigint NOT NULL,
			reward bigint NOT NULL,
			reward_lv integer NOT NULL,
			reward_rate double precision NOT NULL DEFAULT 0,
			data VARCHAR,
		time_send_to_wallet timestamp with time zone DEFAULT NULL,
		from_unix bigint ,
			to_unix bigint,
		UNIQUE (user_id, from_unix, to_unix),
			create_time timestamp
		with
		time zone NOT NULL DEFAULT now(),
		update_time timestamp
		with
		time zone NOT NULL DEFAULT now(),
		CONSTRAINT reward_refer_pkey PRIMARY KEY (id)
		)
	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())

	}

	// jackpot
	_, err = db.ExecContext(ctx, `
				CREATE TABLE IF NOT EXISTS
			public.jackpot (
				id bigint NOT NULL,
				game character varying(128) NOT NULL,
				UNIQUE(game),
				chips bigint NOT NULL DEFAULT 0,
				create_time timestamp
				with
				time zone NOT NULL DEFAULT now(),
				update_time timestamp
				with
				time zone NOT NULL DEFAULT now(),
				CONSTRAINT jackpot_pkey PRIMARY KEY (id)
			)
	`)
	if err != nil {
		logger.Error("Error: %s", err.Error())
	}
	// free game
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS
			public.feegame (
			id bigint NOT NULL,
			user_id character varying(128) NOT NULL,
			game character varying(128) NOT NULL,
			fee bigint NOT NULL DEFAULT 0,
			create_time timestamp
			with
			time zone NOT NULL DEFAULT now(),
			update_time timestamp
			with
			time zone NOT NULL DEFAULT now(),
			CONSTRAINT feegame_pkey PRIMARY KEY (id)
		)
`)
	if err != nil {
		logger.Error("Error: %s", err.Error())
	}

	// jackpot history
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS
			public.jackpot_history (
			id bigint NOT NULL,
			game character varying(128) NOT NULL,
			chips bigint NOT NULL DEFAULT 0,
			metadata character varying(256) NOT NULL,
			create_time timestamp
			with
			time zone NOT NULL DEFAULT now(),
			update_time timestamp
			with
			time zone NOT NULL DEFAULT now(),
			CONSTRAINT jackpot_history_pkey PRIMARY KEY (id)
		)
`)
	if err != nil {
		logger.Error("Error: %s", err.Error())
	}

	ddls := []string{
		`CREATE TABLE IF NOT EXISTS public.gold_statistics (
			id bigserial NOT NULL,
			created_at timestamptz NULL,
			updated_at timestamptz NULL,
			deleted_at timestamptz NULL,
			time_update timestamptz NULL,
			pay int8 NULL,
			promotion int8 NULL,
			match_data bytea NULL,
			ag_cashout int8 NULL,
			ag_bank int8 NULL,
			chips int8 NULL,
			CONSTRAINT gold_statistics_pkey PRIMARY KEY (id)
		);
		CREATE INDEX IF NOT EXISTS idx_gold_statistics_deleted_at ON public.gold_statistics USING btree (deleted_at);`,
	}
	ddls = append(ddls, `CREATE TABLE IF NOT EXISTS public.op_match_details (
		id bigserial NOT NULL,
		created_at timestamptz NULL,
		updated_at timestamptz NULL,
		deleted_at timestamptz NULL,
		game_id int8 NULL,
		game_name text NULL,
		mcb int8 NULL,
		match_id text NULL,
		num_match_played int8 NULL,
		chip_fee int8 NULL,
		date_unix int8 NULL,
		detail jsonb NULL,
		CONSTRAINT op_match_details_pkey PRIMARY KEY (id)
	);
	CREATE INDEX IF NOT EXISTS idx_op_match_details_deleted_at ON public.op_match_details USING btree (deleted_at);`)

	ddls = append(ddls, `CREATE TABLE IF NOT EXISTS public.op_players (
		id bigserial NOT NULL,
		created_at timestamptz NULL,
		updated_at timestamptz NULL,
		deleted_at timestamptz NULL,
		user_id text NULL,
		user_name text NULL,
		game_id int8 NULL,
		game_name text NULL,
		mcb int8 NULL,
		no_bet int8 NULL,
		no_win int8 NULL,
		no_lost int8 NULL,
		chip int8 NULL,
		chip_win int8 NULL,
		chip_lost int8 NULL,
		chip_balance int8 NULL,
		date_unix int8 NULL,
		wallet text NULL,
		CONSTRAINT op_players_pkey PRIMARY KEY (id)
	);
	CREATE INDEX IF NOT EXISTS idx_op_players_deleted_at ON public.op_players USING btree (deleted_at);`)

	ddls = append(ddls, `CREATE TABLE IF NOT EXISTS public.op_match_details (
		id bigserial NOT NULL,
		created_at timestamptz NULL,
		updated_at timestamptz NULL,
		deleted_at timestamptz NULL,
		game_id int8 NULL,
		game_name text NULL,
		mcb int8 NULL,
		match_id text NULL,
		num_match_played int8 NULL,
		chip_fee int8 NULL,
		date_unix int8 NULL,
		detail jsonb NULL,
		CONSTRAINT op_match_details_pkey PRIMARY KEY (id)
	);
	CREATE INDEX IF NOT EXISTS idx_op_match_details_deleted_at ON public.op_match_details USING btree (deleted_at);`)

	// ddls = append(ddls, `
	// ALTER TABLE public.in_app_message ADD COLUMN app_package text NULL;
	// ALTER TABLE public.in_app_message ADD COLUMN game_id text NULL;
	// `)
	// ddls = append(ddls, `
	// 	ALTER TABLE public.cgb_notification ADD COLUMN app_package text NULL;
	// 	ALTER TABLE public.cgb_notification ADD COLUMN game_id text NULL;
	// `)
	ddls = append(ddls, `
	CREATE TABLE IF NOT EXISTS public.bets (
		id bigserial NOT NULL,
		created_at timestamptz NULL,
		game_id int8 NULL,
		mark_unit float8 NULL,
		x_join float8 NULL,
		x_play_now float8 NULL,
		x_leave float8 NULL,
		x_fee float8 NULL,
		new_fee float8 NULL,
		CONSTRAINT bets_pkey PRIMARY KEY (id)
	);
`)

	ddls = append(ddls, `
	CREATE TABLE IF NOT EXISTS public.games (
		id bigserial NOT NULL,
		created_at timestamptz NULL DEFAULT now(),
		code varchar(31) NOT NULL,
		CONSTRAINT games_code_key UNIQUE (code),
		CONSTRAINT games_pkey PRIMARY KEY (id)
	);
`)

	ddls = append(ddls, `
CREATE TABLE IF NOT EXISTS public.rules_lucky (
	id bigserial NOT NULL,
	create_at timestamptz NULL DEFAULT now(),
	game_code varchar(31) NOT NULL,	
	emit_event_at_unix int8 DEFAULT 1,
	deleted_at int8 DEFAULT 0,
	rtp_min int8 NOT NULL DEFAULT 0,
	rtp_max int8 NOT NULL DEFAULT 0,
	mark_min int8 NOT NULL DEFAULT 0,
	mark_max int8 NOT NULL DEFAULT 0,
	vip_min int4 NOT NULL DEFAULT 0,
	vip_max int4 NOT NULL DEFAULT 0,
	win_mark_ratio_min int8 NOT NULL DEFAULT 0,
	win_mark_ratio_max int8 NOT NULL DEFAULT 0,
	re_deal int8 NOT NULL DEFAULT 0
);
`)
	ddls = append(ddls, `
CREATE TABLE IF NOT EXISTS public.users_bot (
	id bigserial NOT NULL,
	user_id varchar(36) NOT NULL,
	game_code varchar(31) NOT NULL
);
`)

	// Bot config table for bot management system
	ddls = append(ddls, `
-- Bot Join Rules Table
CREATE TABLE IF NOT EXISTS public.bot_join_rules (
    id SERIAL PRIMARY KEY,
    game_code VARCHAR(50) NOT NULL,
    min_bet INTEGER NOT NULL,
    max_bet INTEGER NOT NULL,
    min_users INTEGER NOT NULL,
    max_users INTEGER NOT NULL,
    random_time_min INTEGER NOT NULL,
    random_time_max INTEGER NOT NULL,
    join_percent INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Bot Leave Rules Table
CREATE TABLE IF NOT EXISTS public.bot_leave_rules (
    id SERIAL PRIMARY KEY,
    game_code VARCHAR(50) NOT NULL,
    min_bet INTEGER NOT NULL,
    max_bet INTEGER NOT NULL,
    last_result INTEGER NOT NULL,
    leave_percent INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Bot Create Table Rules Table
CREATE TABLE IF NOT EXISTS public.bot_create_table_rules (
    id SERIAL PRIMARY KEY,
    game_code VARCHAR(50) NOT NULL,
    min_bet INTEGER NOT NULL,
    max_bet INTEGER NOT NULL,
    min_active_tables INTEGER NOT NULL,
    max_active_tables INTEGER NOT NULL,
    wait_time_min INTEGER NOT NULL,
    wait_time_max INTEGER NOT NULL,
    retry_wait_min INTEGER NOT NULL,
    retry_wait_max INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Bot Group Rules Table
CREATE TABLE IF NOT EXISTS public.bot_group_rules (
    id SERIAL PRIMARY KEY,
    game_code VARCHAR(50) NOT NULL,
    vip_min INTEGER NOT NULL,
    vip_max INTEGER NOT NULL,
    mcb_min INTEGER NOT NULL,
    mcb_max INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_bot_join_rules_game_active ON public.bot_join_rules(game_code, is_active);
CREATE INDEX IF NOT EXISTS idx_bot_leave_rules_game_active ON public.bot_leave_rules(game_code, is_active);
CREATE INDEX IF NOT EXISTS idx_bot_create_table_rules_game_active ON public.bot_create_table_rules(game_code, is_active);
CREATE INDEX IF NOT EXISTS idx_bot_group_rules_game_active ON public.bot_group_rules(game_code, is_active);
`)
	for _, ddl := range ddls {
		_, err = db.ExecContext(ctx, ddl)
		if err != nil {
			logger.WithField("err", err).Error("ddl failed")
		}
	}
	logger.Info("Done run migration")
}
