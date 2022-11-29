begin;

create table player_team_game_stats_period
(
    id                       uuid                     default gen_random_uuid() not null primary key,
    created_at               timestamp with time zone default now()             not null,
    updated_at               timestamp with time zone,
    game_id                  uuid                                               not null references game (id),
    team_id                  uuid                                               not null references team (id),
    player_id                uuid                                               not null references player (id),
    period                   integer                                            not null,
    time_played_seconds      integer,
    points                   integer,
    assists                  integer,
    turnovers                integer,
    steals                   integer,
    three_pointers_attempted integer,
    three_pointers_made      integer,
    three_point_percentage   numeric(4, 3),
    field_goals_attempted    integer,
    field_goals_made         integer,
    field_goal_percentage    numeric(4, 3),
    free_throws_attempted    integer,
    free_throws_made         integer,
    free_throw_percentage    numeric(4, 3),
    blocks                   integer,
    rebounds_offensive       integer,
    rebounds_defensive       integer,
    rebounds_total           integer,
    fouls_personal           integer,
    plus_minus               integer,
    unique (game_id, team_id, player_id, period)
);

create trigger set_timestamp
    before update
    on player_team_game_stats_period
    for each row
execute procedure trigger_set_timestamp();

commit;