-- name: ListUsers :many
select * from users;

-- name: GetUser :one
select * from users 
where id = $1;

-- name: GetUserByUsername :one
select * from users 
where username = $1;

-- name: CreateUser :one
insert into users (username,password) 
values ($1,$2)
returning *;
