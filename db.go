package main

import (
	"database/sql"
	"fmt"

	"github.com/gempir/go-twitch-irc/v4"
	_ "github.com/lib/pq"
)

var db *sql.DB

func saveChatMessage(message twitch.PrivateMessage) error {
	stmt := "INSERT INTO messages(created_at, username, message) VALUES ($1, $2, $3)"

	_, err := db.Exec(stmt, message.Time, message.User.DisplayName, message.Message)
	return err
}

func init() {

	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		"localhost", "twitch", "twitch", "twitch", 5433, "disable", "Europe/Berlin")

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		panic("failed to connect database")
	}

	initStmt := `
	CREATE TABLE messages (
		id serial primary key not null,
		created_at timestamp without time zone not null default now(),
		username varchar not null,
		message varchar not null
	);
	`

	_, err = db.Exec(initStmt)
	if err != nil {
		panic(err)
	}
}
