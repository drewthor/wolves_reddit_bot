begin;

create table game_referee
(
    id         uuid                     default gen_random_uuid() not null primary key,
    created_at timestamp with time zone default now()             not null,
    updated_at timestamp with time zone,
    game_id    uuid                                               not null references game (id),
    referee_id uuid                                               not null references referee (id),
    assignment text,
    unique (game_id, referee_id)
);

create or replace trigger set_timestamp
    before update
    on game_referee
    for each row
execute procedure trigger_set_timestamp();

commit;