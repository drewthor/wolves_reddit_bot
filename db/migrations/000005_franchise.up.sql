begin;

create table franchise
(
    id         uuid                     default gen_random_uuid() not null primary key,
    created_at timestamp with time zone default now()             not null,
    updated_at timestamp with time zone,
    name       text                                               not null,
    nickname   text                                               not null,
    city       text                                               not null,
    state      text                                               not null,
    country    text                                               not null,
    active     boolean                                            not null
);

create or replace trigger set_timestamp
    before update
    on franchise
    for each row
execute procedure trigger_set_timestamp();

commit;