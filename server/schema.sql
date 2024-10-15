create type visibility as enum ('public', 'private');
create type role as enum ('admin','user', 'guest');

create table if not exists  users (
    id  serial primary key,
    username  text not null unique,
    password  text not null,
    created_at  timestamp  with time zone default current_timestamp
);
create table if not exists workspaces (
    id serial primary key,
    user_id integer references users(id),
    name  text unique,
    tags text,
    view visibility
);
create table if not exists sessions (
    id text primary key,
    user_id integer not null references users(id),
    expires_at timestamp
);
