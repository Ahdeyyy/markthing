package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func newDb(host, user, password, name string) (*sql.DB, error) {

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		host, user, password, name)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	// NOTE: remove this
	// dropTables(db)
	createTables(db)

	return db, nil

}

func dropTables(db *sql.DB) {
	log.Println("dropping tables...")
	stmt := `drop table users cascade;
		drop table workspaces cascade;
		drop table sessions cascade;`
	db.Exec(stmt)

}
func createTables(db *sql.DB) {
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

	_, err := db.Exec(enums)
	if err != nil {
		log.Printf("error creating enum types: %s", err)
	}
	_, err = db.Exec(userTable)
	if err != nil {
		log.Printf("error creating users table: %s", err)
	}
	_, err = db.Exec(sessionTable)
	if err != nil {
		log.Printf("error creating sessions table: %s", err)
	}
	_, err = db.Exec(workspaceTable)
	if err != nil {
		log.Printf("error creating workspaces table: %s", err)
	}

}
