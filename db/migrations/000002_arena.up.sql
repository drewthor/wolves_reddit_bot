begin;

create table arena
(
    id           uuid                     default gen_random_uuid() not null primary key,
    created_at   timestamp with time zone default now()             not null,
    updated_at   timestamp with time zone,
    name         text                                               not null,
    city         text,
    state        text,
    country      text                                               not null,
    nba_arena_id integer                                            not null unique
);

create trigger set_timestamp
    before update
    on arena
    for each row
execute procedure trigger_set_timestamp();

commit;