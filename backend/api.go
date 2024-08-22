package main

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func loginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")

	fail_context := map[string]string{
		"email":         email,
		"password":      password,
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
	email := r.FormValue("email")
	name := r.FormValue("name")
	password_1 := r.FormValue("password_1")
	password_2 := r.FormValue("password_2")

	failContext := map[string]string{
		"email":         email,
		"name":          name,
		"password_1":    password_1,
		"password_2":    password_2,
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