begin;

create table season
(
    id         uuid                     default gen_random_uuid() not null primary key,
    created_at timestamp with time zone default now()             not null,
    updated_at timestamp with time zone,
    start_year integer                                            not null,
    end_year   integer                                            not null,
    league_id  uuid                                               not null references league (id),
    unique (start_year, end_year, league_id)
);

create or replace trigger set_timestamp
    before update
    on season
    for each row
execute procedure trigger_set_timestamp();

commit;