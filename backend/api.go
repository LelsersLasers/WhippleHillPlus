package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	_, sessionID := isLoggedIn(r)
	user, err := userFromEmail(sessionID)

	if err != nil {
		logout(&w, r)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	classes, assignments := allClassesAndAssignments(user.ID)
	data := map[string]interface{}{
		"user":        user,
		"classes":     classes,
		"assignments": assignments,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	loggedIn, _ := isLoggedIn(r)
	if loggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

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

	mutex.Lock()
	defer mutex.Unlock()

	rows, err := db.Query("SELECT password_hash FROM users WHERE email = ?", email)
	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() {
		passwordHash := ""
		rows.Scan(&passwordHash)
		if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil {
			login(&w, r, email)
			return
		}
	}

	fail_context["error_message"] = "Invalid email or password"
	failContextToCookies(&w, fail_context)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func logoutUser(w http.ResponseWriter, r *http.Request) {
	logout(&w, r)
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	loggedIn, _ := isLoggedIn(r)
	if loggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

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
	defer mutex.Unlock()

	rows, err := db.Query("SELECT * FROM users WHERE email = ?", email)
	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

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
	passwordHashStr := string(passwordHash)

	_, err = db.Exec("INSERT INTO users (email, name, password_hash) VALUES (?, ?, ?)", email, name, passwordHashStr)
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

	var data struct {
		Name          string `json:"name"`
		Description    string `json:"description"`
		DueDate        string `json:"due_date"`
		DueTime        string `json:"due_time"`
		AssignedDate   string `json:"assigned_date"`
		Status         string `json:"status"`
		AssignmentType string `json:"type"`
		ClassID        string `json:"class_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	res, err := db.Exec("INSERT INTO assignments (name, description, due_date, due_time, assigned_date, status, type, class_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", data.Name, data.Description, data.DueDate, data.DueTime, data.AssignedDate, data.Status, data.AssignmentType, data.ClassID)
	if err != nil {
		http.Error(w, "Internal server error - failed to insert assignment", http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Internal server error - failed to insert assignment", http.StatusInternalServerError)
		return
	}

	assignment := Assignment{}

	rows, err := db.Query("SELECT * FROM assignments WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

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
	defer mutex.Unlock()

	rows, err := db.Query("SELECT * FROM assignments WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

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

	var data struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		Description    string `json:"description"`
		DueDate        string `json:"due_date"`
		DueTime        string `json:"due_time"`
		AssignedDate   string `json:"assigned_date"`
		Status         string `json:"status"`
		AssignmentType string `json:"type"`
		ClassID        string `json:"class_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	_, err = db.Exec("UPDATE assignments SET name = ?, description = ?, due_date = ?, due_time = ?, assigned_date = ?, status = ?, type = ?, class_id = ? WHERE id = ?", data.Name, data.Description, data.DueDate, data.DueTime, data.AssignedDate, data.Status, data.AssignmentType, data.ClassID, data.ID)

	if err != nil {
		http.Error(w, "Internal server error - failed to update assignment", http.StatusInternalServerError)
		return
	}

	assignment := Assignment{}

	rows, err := db.Query("SELECT * FROM assignments WHERE id = ?", data.ID)
	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

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

	var data struct {
		ID int `json:"id"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	_, err = db.Exec("DELETE FROM assignments WHERE id = ?", data.ID)

	if err != nil {
		http.Error(w, "Internal server error - failed to delete assignment", http.StatusInternalServerError)
		return
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
	defer mutex.Unlock()

	_, err := db.Exec("UPDATE assignments SET status = ? WHERE id = ?", status, id)

	if err != nil {
		http.Error(w, "Internal server error - failed to update assignment status", http.StatusInternalServerError)
		return
	}

	assignment := Assignment{}

	rows, err := db.Query("SELECT * FROM assignments WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

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

	var data struct {
		Name   string `json:"name"`
		UserID string `json:"user_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	res, err := db.Exec("INSERT INTO classes (name, user_id) VALUES (?, ?)", data.Name, data.UserID)
	if err != nil {
		http.Error(w, "Internal server error - failed to insert class", http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Internal server error - failed to get last id", http.StatusInternalServerError)
		return
	}

	class := Class{}

	rows, err := db.Query("SELECT * FROM classes WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

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

	var data struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	_, err = db.Exec("UPDATE classes SET name = ? WHERE id = ?", data.Name, data.ID)
	if err != nil {
		http.Error(w, "Internal server error - failed to update class", http.StatusInternalServerError)
		return
	}

	class := Class{}

	rows, err := db.Query("SELECT * FROM classes WHERE id = ?", data.ID)
	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

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

	var data struct {
		ID int `json:"id"`
	}
	
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	_, err = db.Exec("DELETE FROM classes WHERE id = ?", data.ID)

	if err != nil {
		http.Error(w, "Internal server error - failed to delete class", http.StatusInternalServerError)
		return
	}
}
