begin;

create table team_season
(
    id            uuid                     default gen_random_uuid() not null primary key,
    created_at    timestamp with time zone default now()             not null,
    updated_at    timestamp with time zone,
    team_id       uuid not null references team (id),
    season_id     uuid not null references season (id),
    league_id     uuid not null references league (id),
    conference_id uuid not null references conference (id),
    division_id   uuid not null references division (id),
    unique (team_id, season_id, league_id)
);

create or replace trigger set_timestamp
    before update
    on team_season
    for each row
execute procedure trigger_set_timestamp();

commit;