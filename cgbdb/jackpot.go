package cgbdb

// CREATE TABLE
//   public.jackpot (
//     id bigint NOT NULL,
// 	   game character varying(128) NOT NULL,
//     UNIQUE(game),
//     chips bigint NOT NULL DEFAULT 0,
//     create_time timestamp
//     with
//       time zone NOT NULL DEFAULT now(),
//       update_time timestamp
//     with
//       time zone NOT NULL DEFAULT now()
//   );

// ALTER TABLE
//   public.jackpot
// ADD
//   CONSTRAINT jackpot_pkey PRIMARY KEY (id)

// CREATE TABLE
//   public.jackpot_history (
//     id bigint NOT NULL,
// 	   game character varying(128) NOT NULL,
//     chips bigint NOT NULL DEFAULT 0,
//     metadata string
//     create_time timestamp
//     with
//       time zone NOT NULL DEFAULT now(),
//       update_time timestamp
//     with
//       time zone NOT NULL DEFAULT now()
//   );

// ALTER TABLE
//   public.jackpot_history
// ADD
//   CONSTRAINT jackpot_pkey PRIMARY KEY (id)
