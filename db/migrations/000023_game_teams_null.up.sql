begin;

alter table game alter column home_team_id drop not null;
alter table game alter column away_team_id drop not null;

commit;
