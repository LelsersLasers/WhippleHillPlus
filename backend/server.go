package main

import (
	"database/sql"
	"fmt"
	"os"
	"net/http"
	"sync"
	"time"
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



func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		fmt.Printf("%s - %s %s %s\n", now.Format("2006-01-02 15:04:05"), r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}






func main() {
	db = dbConn()
	createTables(db)
	defer db.Close()

	

	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	loggedHandler := loggingMiddleware(handler)


	fmt.Printf("Server is running on port %d\n", Port)

	addr := fmt.Sprintf(":%d", Port)
	err := http.ListenAndServe(addr, loggedHandler)
	if err != nil {
		fmt.Println("Error starting server: ", err)
	}
}
