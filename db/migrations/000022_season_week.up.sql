begin;

create table season_week
(
    id         uuid                     default gen_random_uuid() not null primary key,
    created_at timestamp with time zone default now()             not null,
    updated_at timestamp with time zone,
    season_id  uuid references season (id),
    start_date timestamp with time zone                           not null,
    end_date   timestamp with time zone                           not null,
    unique (season_id, start_date),
    unique (season_id, end_date)
);

create trigger set_timestamp
    before update
    on season_week
    for each row
execute procedure trigger_set_timestamp();

commit;
