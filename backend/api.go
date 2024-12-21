package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func homeData(w http.ResponseWriter, r *http.Request) {
	loggedIn, username := isLoggedIn(r)
	user, err := userFromUsername(username)

	if !loggedIn || err != nil {
		logout(&w, r)
		data := map[string]interface{}{
			"error": "Not logged in",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
		return
	}

	// classes, assignments := allClassesAndAssignments(user.ID)
	semesters, classes, assignments := allSemestersClassesAndAssignments(user.ID)
	data := map[string]interface{}{
		"user":        user,
		"semesters":   semesters,
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
	username := r.FormValue("username")
	password := r.FormValue("password")

	fail_context := map[string]string{
		"username":      username,
		"password":      password,
		"error_message": "",
	}

	mutex.Lock()
	defer mutex.Unlock()

	rows, err := db.Query("SELECT password_hash FROM users WHERE username = ?", username)
	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() {
		passwordHash := ""
		rows.Scan(&passwordHash)
		if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil {
			login(&w, r, username)
			return
		}
	}

	fail_context["error_message"] = "Invalid username or password"
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
	username := r.FormValue("username")
	name := r.FormValue("name")
	password_1 := r.FormValue("password_1")
	password_2 := r.FormValue("password_2")

	failContext := map[string]string{
		"username":      username,
		"name":          name,
		"password_1":    password_1,
		"password_2":    password_2,
		"error_message": "",
	}

	if username == "" || name == "" || password_1 == "" || password_2 == "" {
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

	rows, err := db.Query("SELECT * FROM users WHERE username = ?", username)
	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() {
		failContext["error_message"] = "Username already in use"
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

	_, err = db.Exec("INSERT INTO users (username, name, password_hash) VALUES (?, ?, ?)", username, name, passwordHashStr)
	if err != nil {
		http.Error(w, "Internal server error - failed to insert user", http.StatusInternalServerError)
		return
	}

	login(&w, r, username)
}

func createAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	var data struct {
		Name           string `json:"name"`
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

func updateAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	var data struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
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

	var data struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	_, err = db.Exec("UPDATE assignments SET status = ? WHERE id = ?", data.Status, data.ID)

	if err != nil {
		http.Error(w, "Internal server error - failed to update assignment status", http.StatusInternalServerError)
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

func createClass(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	var data struct {
		Name       string `json:"name"`
		UserID     string `json:"user_id"`
		SemesterID string `json:"semester_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	res, err := db.Exec("INSERT INTO classes (name, user_id, semester_id) VALUES (?, ?, ?)", data.Name, data.UserID, data.SemesterID)
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
		rows.Scan(&class.ID, &class.Name, &class.UserID, &class.SemesterID)
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
		ID         string `json:"id"`
		Name       string `json:"name"`
		SemesterID string `json:"semester_id"` // Added this line
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	_, err = db.Exec("UPDATE classes SET name = ?, semester_id = ? WHERE id = ?", data.Name, data.SemesterID, data.ID)
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
		rows.Scan(&class.ID, &class.Name, &class.UserID, &class.SemesterID)
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

func createSemester(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	var data struct {
		Name      string `json:"name"`
		UserID    string `json:"user_id"`
		SortOrder string `json:"sort_order"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	res, err := db.Exec("INSERT INTO semesters (name, sort_order, user_id) VALUES (?, ?, ?)", data.Name, data.SortOrder, data.UserID)
	if err != nil {
		http.Error(w, "Internal server error - failed to insert semester", http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Internal server error - failed to get last id", http.StatusInternalServerError)
		return
	}

	semester := Semester{}

	rows, err := db.Query("SELECT * FROM semesters WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&semester.ID, &semester.Name, &semester.SortOrder, &semester.UserID)
	} else {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(semester)
}

func updateSemester(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	var data struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		SortOrder string `json:"sort_order"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	_, err = db.Exec("UPDATE semesters SET name = ?, sort_order = ? WHERE id = ?", data.Name, data.SortOrder, data.ID)
	if err != nil {
		http.Error(w, "Internal server error - failed to update semester", http.StatusInternalServerError)
		return
	}

	semester := Semester{}

	rows, err := db.Query("SELECT * FROM semesters WHERE id = ?", data.ID)
	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&semester.ID, &semester.Name, &semester.SortOrder, &semester.UserID)
	} else {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(semester)
}

func deleteSemester(w http.ResponseWriter, r *http.Request) {
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

	_, err = db.Exec("DELETE FROM semesters WHERE id = ?", data.ID)

	if err != nil {
		http.Error(w, "Internal server error - failed to delete semester", http.StatusInternalServerError)
		return
	}
}
