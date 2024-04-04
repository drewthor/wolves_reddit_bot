begin;

create table team
(
    id             uuid                     default gen_random_uuid() not null primary key,
    created_at     timestamp with time zone default now()             not null,
    updated_at     timestamp with time zone,
    name           text                                               not null,
    nickname       text                                               not null,
    city           text,
    city_alternate text,
    state          text,
    country        text,
    franchise_id   uuid                                               references franchise (id) not null,
    nba_url_name   text                                               not null,
    nba_short_name text                                               not null,
    nba_team_id    integer                                            not null unique
);

create or replace trigger set_timestamp
    before update
    on team
    for each row
execute procedure trigger_set_timestamp();

commit;