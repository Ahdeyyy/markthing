// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package repository

import (
	"context"
)

const createUser = `-- name: CreateUser :one
insert into users (username,password) 
values ($1,$2)
returning id, username, password, created_at
`

type CreateUserParams struct {
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"password"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.Username, arg.Password)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Password,
		&i.CreatedAt,
	)
	return i, err
}

const getUser = `-- name: GetUser :one
select id, username, password, created_at from users 
where id = $1
`

func (q *Queries) GetUser(ctx context.Context, id int32) (User, error) {
	row := q.db.QueryRow(ctx, getUser, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Password,
		&i.CreatedAt,
	)
	return i, err
}

const getUserByUsername = `-- name: GetUserByUsername :one
select id, username, password, created_at from users 
where username = $1
`

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByUsername, username)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Password,
		&i.CreatedAt,
	)
	return i, err
}

const listUsers = `-- name: ListUsers :many
select id, username, password, created_at from users
`

func (q *Queries) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := q.db.Query(ctx, listUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.Password,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
