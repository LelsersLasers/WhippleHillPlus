package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// const Port = 3003
const Port = 8100
const DbPath = "./database.db"

const SessionIdCookieName = "session_id"
const SessionTimeout = 2 * 7 * 24 * time.Hour

const ContextFailCookieNameBase = "context_fail_"
const ContextFailCookieTimeout = 5 * time.Second

const SvelteDir = "./../frontend/public"

var (
	db    *sql.DB
	mutex sync.Mutex
)

func main() {
	db = dbConn()
	createTables(db)
	defer db.Close()

	handler := http.NewServeMux()

	fileServer := http.FileServer(http.Dir(SvelteDir))
	handler.Handle("/", checkLogin(fileServer))

	handler.HandleFunc("/home_data", homePage)

	handler.HandleFunc("/login", loginPage)
	handler.HandleFunc("/register", registerPage)

	handler.HandleFunc("/login_user", loginUser)
	handler.HandleFunc("/logout_user", logoutUser)
	handler.HandleFunc("/register_user", registerUser)

	handler.HandleFunc("/create_assignment", createAssignment)
	handler.HandleFunc("/update_assignment", updateAssignment)
	handler.HandleFunc("/delete_assignment", deleteAssignment)
	handler.HandleFunc("/status_assignment", statusAssignment)

	handler.HandleFunc("/create_class", createClass)
	handler.HandleFunc("/update_class", updateClass)
	handler.HandleFunc("/delete_class", deleteClass)

	handler.HandleFunc("/create_semester", createSemester)
	handler.HandleFunc("/update_semester", updateSemester)
	handler.HandleFunc("/delete_semester", deleteSemester)

	middlewareHandler := loggingMiddleware(corsMiddleware(handler))

	fmt.Printf("Server is running on port %d\n", Port)

	addr := fmt.Sprintf("[::]:%d", Port)
	err := http.ListenAndServe(addr, middlewareHandler)
	if err != nil {
		fmt.Println("Error starting server: ", err)
	}
}
