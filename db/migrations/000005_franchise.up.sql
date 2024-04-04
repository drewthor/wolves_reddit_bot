begin;

create table franchise
(
    id                  uuid                     default gen_random_uuid() not null primary key,
    created_at          timestamp with time zone default now()             not null,
    updated_at          timestamp with time zone,
    name                text                                               not null,
    nickname            text                                               not null,
    city                text                                               not null,
    state               text                                               not null,
    country             text                                               not null,
    league_id           uuid references league (id)                        not null,
    nba_team_id         integer                                            not null unique,
    years               integer                                            not null,
    games               integer                                            not null,
    wins                integer                                            not null,
    losses              integer                                            not null,
    playoff_appearances integer                                            not null,
    division_titles     integer                                            not null,
    conference_titles   integer                                            not null,
    league_titles       integer                                            not null,
    active              boolean                                            not null
);

create or replace trigger set_timestamp
    before update
    on franchise
    for each row
execute procedure trigger_set_timestamp();

commit;