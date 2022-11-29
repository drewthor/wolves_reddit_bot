begin;

create table player_position
(
    id          uuid                     default gen_random_uuid() not null primary key,
    created_at  timestamp with time zone default now()             not null,
    updated_at  timestamp with time zone,
    player_id   uuid                                               not null references player (id),
    position_id uuid                                               not null references position (id),
    priority    integer                                            not null,
    unique (player_id, position_id, priority)
);

create trigger set_timestamp
    before update
    on player_position
    for each row
execute procedure trigger_set_timestamp();

commit;