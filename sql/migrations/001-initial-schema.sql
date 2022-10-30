create table notification (
    id text not null primary key,
    email_destination text not null,
    sms_destination text not null,
    content text,
    schedule timestamptz,
    created_at timestamptz not null default now(),
    modified_at timestamptz not null default now(),
    deleted_at timestamptz,
    delivered_at timestamptz,
    submitted_at timestamptz
);

create index if not exists notification_schedule on notification(delivered_at, deleted_at, schedule);

create or replace function update_modified_on() returns trigger as $$
begin
    NEW.modified_at := now();
    return NEW;
end;
$$ language plpgsql;

create trigger account_modified_on
    before update or insert
    on notification
    for each row
execute procedure update_modified_on();
