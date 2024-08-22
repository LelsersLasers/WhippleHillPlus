package main

import (
	"database/sql"
	"fmt"
	"os"
	"net/http"
	"html/template"
	"sync"
	"time"
	// "encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)


const Port = 8080
const DbPath = "./database.db"

const SessionIdCookieName = "session_id"
const SessionTimeout = 24 * time.Hour

const ContextFailCookieNameBase = "context_fail_"
const ContextFailCookieTimeout = 5 * time.Second

var (
	db *sql.DB
	mutex sync.Mutex
)


type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	PasswordHash string `json:"password_hash"`
}

type Assignment struct {
	ID           int    `json:"id"`
	Name         string `json:"title"`
	Description  string `json:"description"`
	DueDate      string `json:"due_date"`
	DueTime      string `json:"due_time"`
	AssignedDate string `json:"assigned_date"`
	Status 	     string `json:"status"`
	Type		 string `json:"type"`
	ClassID      int    `json:"class_id"`
}




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
func userFromEmail(email string) (User, error) {
	user := User{}

	mutex.Lock()
	rows, err := db.Query("SELECT * FROM users WHERE email = ?", email)
	mutex.Unlock()

	if err != nil {
		return user, err
	}

	if rows.Next() {
		rows.Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash)
		return user, nil
	}

	return user, fmt.Errorf("User not found")
}

func login(w *http.ResponseWriter, r *http.Request, email string) {
	// Already verified that login info is correct
	sessionID := email
	http.SetCookie(*w, &http.Cookie{
		Name: SessionIdCookieName,
		Value: sessionID,
		Expires: time.Now().Add(SessionTimeout),
		Path: "/",
	})
	http.Redirect(*w, r, "/", http.StatusSeeOther)
}

func logout(w *http.ResponseWriter, r *http.Request) {
	http.SetCookie(*w, &http.Cookie{
		Name: SessionIdCookieName,
		Value: "",
		Expires: time.Now(),
	})

	http.Redirect(*w, r, "/login", http.StatusSeeOther)
}

func failContextToCookies(w *http.ResponseWriter, failContext map[string]string) {
	for key, value := range failContext {
		http.SetCookie(*w, &http.Cookie{
			Name: ContextFailCookieNameBase + key,
			Value: value,
			Expires: time.Now().Add(ContextFailCookieTimeout),
		})
	}
}
func cookiesToFailContext(failContext map[string]string, w *http.ResponseWriter, r *http.Request) map[string]string {
	for key := range failContext {
		context_fail_cookie, err := r.Cookie(ContextFailCookieNameBase + key)
		if err == nil {
			failContext[key] = context_fail_cookie.Value
			http.SetCookie(*w, &http.Cookie{
				Name: ContextFailCookieNameBase + key,
				Value: "",
				Expires: time.Now(),
			})
		}
	}

	return failContext
}

func allAssignments(user_id int) []Assignment {
	assignments := []Assignment{}
	
	mutex.Lock()
	rows, err := db.Query("SELECT * FROM assignments WHERE user_id = ?", user_id)
	mutex.Unlock()

	if err != nil {
		return assignments
	}

	for rows.Next() {
		assignment := Assignment{}
		rows.Scan(&assignment.ID, &assignment.Name, &assignment.Description, &assignment.DueDate, &assignment.DueTime, &assignment.AssignedDate, &assignment.Status, &assignment.Type, &assignment.ClassID)
		assignments = append(assignments, assignment)
	}
	
	return assignments
}


// -------------------------------------------------------------------------- //
func loginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	r.ParseForm()
	email    := r.FormValue("email")
	password := r.FormValue("password")

	fail_context := map[string]string{
		"email": email,
		"password": password,
		"error_message": "",
	}

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
		login(&w, r, email)
	} else {
		fail_context["error_message"] = "Invalid email or password"
		failContextToCookies(&w, fail_context)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func logoutUser(w http.ResponseWriter, r *http.Request) {
	logout(&w, r)
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	r.ParseForm()
	email      := r.FormValue("email")
	name       := r.FormValue("name")
	password_1 := r.FormValue("password_1")
	password_2 := r.FormValue("password_2")

	failContext := map[string]string{
		"email": email,
		"name": name,
		"password_1": password_1,
		"password_2": password_2,
		"error_message": "",
	}

	if email == "" || name == "" || password_1 == "" || password_2 == "" {
		failContext["error_message"] = "All fields are required"
		failContextToCookies(&w, failContext)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	if password_1 != password_2 {
		failContext["error_message"] = "Passwords do not match"
		failContextToCookies(&w, failContext)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
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
		failContext["error_message"] = "Email already in use"
		failContextToCookies(&w, failContext)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
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

	login(&w, r, email)
}
// -------------------------------------------------------------------------- //


// -------------------------------------------------------------------------- //
func homePage(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionIdCookieName)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessionID := cookie.Value

	user, err := userFromEmail(sessionID)

	if err != nil {
		logout(&w, r)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	assignments := allAssignments(user.ID)
	templ := template.Must(template.ParseFiles("templates/home.html"))
	templ.Execute(w, assignments)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	failContext := map[string]string{
		"email": "",
		"password": "",
		"error_message": "",
	}
	failContext = cookiesToFailContext(failContext, &w, r)

	templ := template.Must(template.ParseFiles("templates/login.html"))
	templ.Execute(w, failContext)
}

func registerPage(w http.ResponseWriter, r *http.Request) {
	failContext := map[string]string{
		"email": "",
		"name": "",
		"password_1": "",
		"password_2": "",
		"error_message": "",
	}
	failContext = cookiesToFailContext(failContext, &w, r)

	templ := template.Must(template.ParseFiles("templates/register.html"))
	templ.Execute(w, failContext)
}
// -------------------------------------------------------------------------- //



// -------------------------------------------------------------------------- //
func main() {
	db = dbConn()
	createTables(db)
	defer db.Close()


	handler := http.NewServeMux()

	handler.HandleFunc("/", homePage)
	handler.HandleFunc("/login", loginPage)
	handler.HandleFunc("/register", registerPage)


	handler.HandleFunc("/login_user", loginUser)
	handler.HandleFunc("/logout_user", logoutUser)
	handler.HandleFunc("/register_user", registerUser)

	loggedHandler := loggingMiddleware(handler)


	fmt.Printf("Server is running on port %d\n", Port)

	addr := fmt.Sprintf(":%d", Port)
	err := http.ListenAndServe(addr, loggedHandler)
	if err != nil {
		fmt.Println("Error starting server: ", err)
	}
}
