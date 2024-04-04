begin;

create table play_by_play
(
    id         uuid                     default gen_random_uuid() not null primary key,
    created_at timestamp with time zone default now()             not null,
    updated_at timestamp with time zone,
    game_id  uuid references game (id) not null,
    team_id uuid references team(id),
    player_id uuid references player(id),
    period integer not null,
    action_number integer not null,

    unique (game_id, action_number),
);

create or replace trigger set_timestamp
    before update
    on play_by_play
    for each row
execute procedure trigger_set_timestamp();

commit;
