begin;

create function trigger_set_timestamp() returns trigger
    language plpgsql
as
$$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$;

commit;