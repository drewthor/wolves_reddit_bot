begin;

create table referee
(
    id             uuid                     default gen_random_uuid() not null primary key,
    created_at     timestamp with time zone default now(),
    updated_at     timestamp with time zone,
    first_name     text,
    last_name      text,
    nba_referee_id integer unique,
    jersey_number  integer
);

create trigger set_timestamp
    before update
    on referee
    for each row
execute procedure trigger_set_timestamp();

commit;