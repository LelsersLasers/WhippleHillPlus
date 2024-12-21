package main

import (
	"database/sql"
	"os"
)

func dbConn() (db *sql.DB) {
	db, err := sql.Open("sqlite3", DbPath)
	if err != nil {
		panic(err.Error())
	}
	return db
}

func createTables(db *sql.DB) {
	body, err := os.ReadFile("create_tables.sql")
	if err != nil {
		panic(err.Error())
	}

	query := string(body)

	dbMutex.Lock()
	_, err = db.Exec(query)
	dbMutex.Unlock()

	if err != nil {
		panic(err.Error())
	}
}
