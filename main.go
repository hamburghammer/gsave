package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// TABLENAME name of the table for the application
const TABLENAME = "tweet"

func main() {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	err = tableCleanup(db)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = createTable(db)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = insertInto(db)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = selectAll(db)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func createTable(db *sql.DB) error {
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS tweet ( 'id' INTEGER PRIMARY KEY AUTOINCREMENT, 'username' VARCHAR(64) NULL)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	return nil
}

func insertInto(db *sql.DB) error {
	usernames := []string{"admin", "user", "alice", "bob"}
	stmt, err := db.Prepare("INSERT INTO tweet (username) VALUES (?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, username := range usernames {
		_, err = stmt.Exec(username)
		if err != nil {
			return err
		}
	}

	return nil
}

func selectAll(db *sql.DB) error {
	stmt, err := db.Prepare("SELECT * FROM tweet")
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		id       int
		username string
	)
	for rows.Next() {
		err := rows.Scan(&id, &username)
		if err != nil {
			log.Println(err)
		}

		log.Printf("ID: %d, username: %s", id, username)
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	return nil

}

func tableCleanup(db *sql.DB) error {
	stmt, err := db.Prepare("DROP TABLE IF EXISTS tweet")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	return nil
}
