package main

import (
	"database/sql"
	"fmt"
	"os"
	"net/http"
	"sync"
	_ "github.com/mattn/go-sqlite3"
)


const Port = 8080
const DbPath = "./database.db"

var (
	db *sql.DB
	mutex sync.Mutex
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

	mutex.Lock()
	_, err = db.Exec(query)
	mutex.Unlock()
	
	if err != nil {
		panic(err.Error())
	}
}







func main() {
	db = dbConn()
	createTables(db)
	defer db.Close()

	


	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	fmt.Printf("Server is running on port %d\n", Port)

	addr := fmt.Sprintf(":%d", Port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("Error starting server: ", err)
	}
}
