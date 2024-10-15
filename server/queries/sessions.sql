-- name: GetSession :one
select sqlc.embed(sessions), sqlc.embed(users) from sessions 
inner join users on users.id = sessions.user_id where sessions.id = $1;

-- name: CreateSession :one
insert into sessions (id,user_id, expires_at) 
values ($1,$2, $3)
returning *;

-- name: DeleteSession :exec
delete from sessions where id = $1;

-- name: UpdateSessionExpiresAt :exec
update sessions set expires_at =  $1  where id = $2;
