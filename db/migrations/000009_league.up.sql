begin;

create table league
(
    id            uuid                     default gen_random_uuid() not null primary key,
    created_at    timestamp with time zone default now()             not null,
    updated_at    timestamp with time zone,
    name          text                                               not null,
    nba_league_id integer                                            unique
);

create or replace trigger set_timestamp
    before update
    on league
    for each row
execute procedure trigger_set_timestamp();

commit;