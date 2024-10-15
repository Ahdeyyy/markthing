package store

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func NewConn(host, user, password, name string) (*pgx.Conn, error) {

	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		user, password, host, name)

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	// NOTE: remove this
	// dropTables(db)
	createTables(conn)

	return conn, nil

}

func dropTables(conn *pgx.Conn) {
	log.Println("dropping tables...")
	stmt := `drop table users cascade;
		drop table workspaces cascade;
		drop table sessions cascade;`
	conn.Exec(context.Background(), stmt)

}
func createTables(conn *pgx.Conn) {
	enums := `
	create type visibility as enum ('public', 'private');
	create type role as enum ('admin','user', 'guest');
	`
	userTable := `create table if not exists  users (
		id  serial primary key,
		username  text not null unique,
		password  text not null,
		created_at  timestamp  with time zone default current_timestamp
	);`

	workspaceTable := `
	create table if not exists workspaces (
		id serial primary key,
		user_id integer references users(id),
		name  text unique,
		tags text,
		view visibility
	);`

	sessionTable := `create table if not exists sessions (
		id text primary key,
		user_id integer not null references users(id),
		expires_at timestamp
	)`

	_, err := conn.Exec(context.Background(), enums)
	if err != nil {
		log.Printf("error creating enum types: %s", err)
	}
	_, err = conn.Exec(context.Background(), userTable)
	if err != nil {
		log.Printf("error creating users table: %s", err)
	}
	_, err = conn.Exec(context.Background(), sessionTable)
	if err != nil {
		log.Printf("error creating sessions table: %s", err)
	}
	_, err = conn.Exec(context.Background(), workspaceTable)
	if err != nil {
		log.Printf("error creating workspaces table: %s", err)
	}

}
