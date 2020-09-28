package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		fmt.Println(err)
	}

	defer db.Close()
	createTable("test", db)
	if err != nil {
		fmt.Println(err)
	}
}

func createTable(name string, db *sql.DB) error {
	stmt, err := db.Prepare("CREATE TABLE ? ( 'id' INTEGER PRIMARY KEY AUTOINCREMENT, 'username' VARCHAR(64) NULL)")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(name)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	fmt.Println(rows)
	return nil
}
