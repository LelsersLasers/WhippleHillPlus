package main

import (
	"database/sql"
	"fmt"
	"os"
	"net/http"
	"sync"
	"time"
	// "encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)


const Port = 8080
const DbPath = "./database.db"

const SessionCookieName = "session_id"
const SessionTimeout = 24 * time.Hour

var (
	db *sql.DB
	mutex sync.Mutex
)




// -------------------------------------------------------------------------- //
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
// -------------------------------------------------------------------------- //


// -------------------------------------------------------------------------- //
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		fmt.Printf("%s - %s %s %s\n", now.Format("2006-01-02 15:04:05"), r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}
// -------------------------------------------------------------------------- //


// -------------------------------------------------------------------------- //
func login(w http.ResponseWriter, r *http.Request, email string) {
	// Already verified that login info is correct
	sessionID := email
	http.SetCookie(w, &http.Cookie{
		Name: SessionCookieName,
		Value: sessionID,
		Expires: time.Now().Add(SessionTimeout),
		Path: "/",
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}


// -------------------------------------------------------------------------- //
func loginRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		email    := r.FormValue("email")
		password := r.FormValue("password")

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Internal server error - failed to hash password", http.StatusInternalServerError)
			return
		}

		mutex.Lock()
		rows, err := db.Query("SELECT * FROM users WHERE email = ? AND password_hash = ?", email, passwordHash)
		mutex.Unlock()

		if err != nil {
			http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
			return
		}

		if rows.Next() {
			login(w, r, email)
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	}
}

func logoutRoute(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: SessionCookieName,
		Value: "",
		Expires: time.Now(),
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func registerRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		email      := r.FormValue("email")
		name       := r.FormValue("name")
		password_1 := r.FormValue("password_1")
		password_2 := r.FormValue("password_2")

		if email == "" || name == "" || password_1 == "" || password_2 == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			return
		}

		if password_1 != password_2 {
			http.Error(w, "Passwords do not match", http.StatusBadRequest)
			return
		}

		mutex.Lock()
		rows, err := db.Query("SELECT * FROM users WHERE email = ?", email)
		mutex.Unlock()

		if err != nil {
			http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
			return
		}

		if rows.Next() {
			http.Error(w, "User with email already exists", http.StatusBadRequest)
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password_1), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Internal server error - failed to hash password", http.StatusInternalServerError)
			return
		}

		mutex.Lock()
		_, err = db.Exec("INSERT INTO users (email, name, password_hash) VALUES (?, ?, ?)", email, name, passwordHash)
		mutex.Unlock()

		if err != nil {
			http.Error(w, "Internal server error - failed to insert user", http.StatusInternalServerError)
			return
		}

		login(w, r, email)
	}
}
// -------------------------------------------------------------------------- //


// -------------------------------------------------------------------------- //
func main() {
	db = dbConn()
	createTables(db)
	defer db.Close()

	

	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})
	handler.HandleFunc("/login", loginRoute)
	handler.HandleFunc("/logout", logoutRoute)
	handler.HandleFunc("/register", registerRoute)

	loggedHandler := loggingMiddleware(handler)


	fmt.Printf("Server is running on port %d\n", Port)

	addr := fmt.Sprintf(":%d", Port)
	err := http.ListenAndServe(addr, loggedHandler)
	if err != nil {
		fmt.Println("Error starting server: ", err)
	}
}
