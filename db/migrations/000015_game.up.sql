begin;

create table game
(
    id                                  uuid                     default gen_random_uuid() not null primary key,
    created_at                          timestamp with time zone default now()             not null,
    updated_at                          timestamp with time zone,
    home_team_id                        uuid                                               not null references team (id),
    away_team_id                        uuid                                               not null references team (id),
    game_status_id                      uuid                                               not null references game_status (id),
    arena_id                            uuid                                               references arena (id),
    attendance                          integer,
    season_id                           uuid                                               references season (id),
    season_stage_id                     uuid                                               not null references season_stage (id),
    period                              integer,
    period_time_remaining_tenth_seconds integer,
    duration_seconds                    integer,
    start_time                          timestamp with time zone                           not null,
    end_time                            timestamp with time zone,
    home_team_points                    integer,
    away_team_points                    integer,
    sellout                             boolean,
    regulation_periods                  integer,
    nba_game_id                         text                                               not null unique
);

create or replace trigger set_timestamp
    before update
    on game
    for each row
execute procedure trigger_set_timestamp();

commit;