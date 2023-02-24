begin;

create table player
(
    id               uuid                     default gen_random_uuid() not null primary key,
    created_at       timestamp with time zone default now()             not null,
    updated_at       timestamp with time zone,
    first_name       text                                               not null,
    last_name        text                                               not null,
    birthdate        date,
    height_feet      integer,
    height_inches    integer,
    height_meters    numeric(3, 2),
    weight_pounds    integer,
    weight_kilograms numeric(4, 1),
    jersey_number    smallint,
    positions        text[],
    currently_in_nba boolean                                            not null,
    years_pro        integer                                            not null,
    nba_debut_year   smallint,
    nba_player_id    integer                                            not null unique,
    country          text
);

create or replace trigger set_timestamp
    before update
    on player
    for each row
execute procedure trigger_set_timestamp();

commit;