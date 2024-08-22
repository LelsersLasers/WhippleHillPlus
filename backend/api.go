package main

import (
	"encoding/json"
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


func createAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	
	r.ParseForm()
	title := r.FormValue("title")
	description := r.FormValue("description")
	dueDate := r.FormValue("due_date")
	dueTime := r.FormValue("due_time")
	assignedDate := r.FormValue("assigned_date")
	status := r.FormValue("status")
	assignmentType := r.FormValue("type")
	classID := r.FormValue("class_id")

	mutex.Lock()
	res, err := db.Exec("INSERT INTO assignments (title, description, due_date, due_time, assigned_date, status, type, class_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", title, description, dueDate, dueTime, assignedDate, status, assignmentType, classID)
	if err != nil {
		http.Error(w, "Internal server error - failed to insert assignment", http.StatusInternalServerError)
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Internal server error - failed to insert assignment", http.StatusInternalServerError)
	}
	mutex.Unlock()

	assignment := Assignment{}

	mutex.Lock()
	rows, err := db.Query("SELECT * FROM assignments WHERE id = ?", id)
	mutex.Unlock()

	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

	if rows.Next() {
		rows.Scan(&assignment.ID, &assignment.Name, &assignment.Description, &assignment.DueDate, &assignment.DueTime, &assignment.AssignedDate, &assignment.Status, &assignment.Type, &assignment.ClassID)
	} else {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assignment)
}

func getAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		return
	}

	id := r.URL.Query().Get("id")

	assignment := Assignment{}

	mutex.Lock()
	rows, err := db.Query("SELECT * FROM assignments WHERE id = ?", id)
	mutex.Unlock()

	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

	if rows.Next() {
		rows.Scan(&assignment.ID, &assignment.Name, &assignment.Description, &assignment.DueDate, &assignment.DueTime, &assignment.AssignedDate, &assignment.Status, &assignment.Type, &assignment.ClassID)
	} else {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assignment)
}

func updateAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	r.ParseForm()
	id := r.FormValue("id")
	title := r.FormValue("title")
	description := r.FormValue("description")
	dueDate := r.FormValue("due_date")
	dueTime := r.FormValue("due_time")
	assignedDate := r.FormValue("assigned_date")
	status := r.FormValue("status")
	assignmentType := r.FormValue("type")
	classID := r.FormValue("class_id")

	mutex.Lock()
	_, err := db.Exec("UPDATE assignments SET title = ?, description = ?, due_date = ?, due_time = ?, assigned_date = ?, status = ?, type = ?, class_id = ? WHERE id = ?", title, description, dueDate, dueTime, assignedDate, status, assignmentType, classID, id)
	mutex.Unlock()

	if err != nil {
		http.Error(w, "Internal server error - failed to update assignment", http.StatusInternalServerError)
	}

	assignment := Assignment{}

	mutex.Lock()
	rows, err := db.Query("SELECT * FROM assignments WHERE id = ?", id)
	mutex.Unlock()

	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

	if rows.Next() {
		rows.Scan(&assignment.ID, &assignment.Name, &assignment.Description, &assignment.DueDate, &assignment.DueTime, &assignment.AssignedDate, &assignment.Status, &assignment.Type, &assignment.ClassID)
	} else {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assignment)
}

func deleteAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	r.ParseForm()
	id := r.FormValue("id")

	mutex.Lock()
	_, err := db.Exec("DELETE FROM assignments WHERE id = ?", id)
	mutex.Unlock()

	if err != nil {
		http.Error(w, "Internal server error - failed to delete assignment", http.StatusInternalServerError)
	}
}

func statusAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	r.ParseForm()
	id := r.FormValue("id")
	status := r.FormValue("status")

	mutex.Lock()
	_, err := db.Exec("UPDATE assignments SET status = ? WHERE id = ?", status, id)
	mutex.Unlock()

	if err != nil {
		http.Error(w, "Internal server error - failed to update assignment status", http.StatusInternalServerError)
	}

	assignment := Assignment{}

	mutex.Lock()
	rows, err := db.Query("SELECT * FROM assignments WHERE id = ?", id)
	mutex.Unlock()

	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

	if rows.Next() {
		rows.Scan(&assignment.ID, &assignment.Name, &assignment.Description, &assignment.DueDate, &assignment.DueTime, &assignment.AssignedDate, &assignment.Status, &assignment.Type, &assignment.ClassID)
	} else {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assignment)
}


func createClass(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	r.ParseForm()
	name := r.FormValue("name")
	userID := r.FormValue("user_id")

	mutex.Lock()
	res, err := db.Exec("INSERT INTO classes (name, user_id) VALUES (?, ?)", name, userID)
	if err != nil {
		http.Error(w, "Internal server error - failed to insert class", http.StatusInternalServerError)
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Internal server error - failed to insert class", http.StatusInternalServerError)
	}
	mutex.Unlock()

	class := Class{}

	mutex.Lock()
	rows, err := db.Query("SELECT * FROM classes WHERE id = ?", id)
	mutex.Unlock()

	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

	if rows.Next() {
		rows.Scan(&class.ID, &class.Name, &class.UserID)
	} else {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(class)
}

func updateClass(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	r.ParseForm()
	id := r.FormValue("id")
	name := r.FormValue("name")
	// userID := r.FormValue("user_id")

	mutex.Lock()
	_, err := db.Exec("UPDATE classes SET name = ? WHERE id = ?", name, id)
	mutex.Unlock()

	if err != nil {
		http.Error(w, "Internal server error - failed to update class", http.StatusInternalServerError)
	}

	class := Class{}

	mutex.Lock()
	rows, err := db.Query("SELECT * FROM classes WHERE id = ?", id)
	mutex.Unlock()

	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

	if rows.Next() {
		rows.Scan(&class.ID, &class.Name, &class.UserID)
	} else {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(class)
}

func deleteClass(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	r.ParseForm()
	id := r.FormValue("id")

	mutex.Lock()
	_, err := db.Exec("DELETE FROM classes WHERE id = ?", id)
	mutex.Unlock()

	if err != nil {
		http.Error(w, "Internal server error - failed to delete class", http.StatusInternalServerError)
	}
}