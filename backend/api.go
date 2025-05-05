package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func homeData(w http.ResponseWriter, r *http.Request) {
	loggedIn, username := isLoggedIn(r)

	dbMutex.Lock()
	user, err := userFromUsername(username)
	dbMutex.Unlock()

	if !loggedIn || err != nil {
		logout(&w, r, false)
		data := map[string]interface{}{
			"error": "Not logged in",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
		return
	}

	semesters, classes, assignments := allSemestersClassesAndAssignments(user.ID)
	data := map[string]interface{}{
		"user_name":   user.Name,
		"semesters":   semesters,
		"classes":     classes,
		"assignments": assignments,
	}
	if user.ICSLink != "" {
		data["ics_link"] = user.ICSLink
	}
	if user.Timezone != "" {
		data["timezone"] = user.Timezone
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func icsUpdateTimezoneHandler(w http.ResponseWriter, r *http.Request) {
	loggedIn, username := isLoggedIn(r)
	if !loggedIn {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()
	var data struct {
		Timezone string `json:"timezone"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	_, err = db.Exec("UPDATE users SET timezone = ? WHERE username = ?", data.Timezone, username)
	if err != nil {
		http.Error(w, "Internal server error - failed to update timezone", http.StatusInternalServerError)
		return
	}
	var response struct {
		Timezone string `json:"timezone"`
	}
	response.Timezone = data.Timezone
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func generateICSHandler(w http.ResponseWriter, r *http.Request) {
	loggedIn, username := isLoggedIn(r)
	if !loggedIn {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	icsLink := uuid.New().String()
	_, err := db.Exec("UPDATE users SET ics_link = ? WHERE username = ?", icsLink, username)
	if err != nil {
		http.Error(w, "Internal server error - failed to update ICS link", http.StatusInternalServerError)
		return
	}

	var data struct {
		ICSLink string `json:"ics_link"`
	}
	data.ICSLink = icsLink
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}


func icsHandler(w http.ResponseWriter, r *http.Request) {
	// Path: /ics/<option>/<uuid>.ics
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 4 || pathParts[3] == "" {
		errStr := fmt.Sprintf("Invalid ICS URL: %s", r.URL.Path)
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}

	uuidParts := strings.Split(pathParts[3], ".")
	uuid := uuidParts[0]

	userID, timezone, err := getUserDataFromUUID(uuid)
	if err != nil {
		errStr := fmt.Sprintf("Invalid UUID: %s (%s)", uuid, err.Error())
		http.Error(w, errStr, http.StatusUnauthorized)
		return
	}

	// OPTIONS: 0 = all, 1 = not started, 2 = in progress, 3 = completed
	option := pathParts[2]
	if option != "0" && option != "1" && option != "2" && option != "3" {
		errStr := fmt.Sprintf("Invalid option: %s (%s)", option, r.URL.Path)
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}
	options := map[string]string{"1": "Not Started", "2": "In Progress", "3": "Completed"}
	names := map[string]string{
		"0": "WH+ Assignments (All)",
		"1": "WH+: Not Started",
		"2": "WH+: In Progress",
		"3": "WH+: Completed",
	}

	_, classes, all_assignments := allSemestersClassesAndAssignments(userID)
	assignments := []Assignment{}
	for _, assignment := range all_assignments {
		if option == "0" || assignment.Status == options[option] {
			assignments = append(assignments, assignment)
		}
	}

	icsData := generateICS(classes, assignments, timezone, names[option])

	w.Header().Set("Content-Type", "text/calendar")
	w.Header().Set("Content-Disposition", "attachment; filename=assignments.ics")
	fmt.Fprint(w, icsData)
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

	dbMutex.Lock()
	defer dbMutex.Unlock()

	rows, err := db.Query("SELECT password_hash FROM users WHERE username = ?", username)
	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() {
		passwordHash := ""
		rows.Scan(&passwordHash)
		rows.Close()
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
	logout(&w, r, true)
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

	dbMutex.Lock()
	defer dbMutex.Unlock()

	rows, err := db.Query("SELECT * FROM users WHERE username = ?", username)
	if err != nil {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}

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

	res, err := db.Exec("INSERT INTO users (username, name, password_hash) VALUES (?, ?, ?)", username, name, passwordHashStr)
	if err != nil {
		http.Error(w, "Internal server error - failed to insert user", http.StatusInternalServerError)
		return
	}
	rows.Close()

	// Also create a default semester
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Internal server error - failed to get last id", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO semesters (name, sort_order, user_id) VALUES (?, ?, ?)", DefaultSemesterName, 1, id)
	if err != nil {
		http.Error(w, "Internal server error - failed to insert semester", http.StatusInternalServerError)
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

	dbMutex.Lock()
	defer dbMutex.Unlock()

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

	dbMutex.Lock()
	defer dbMutex.Unlock()

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

	dbMutex.Lock()
	defer dbMutex.Unlock()

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

	dbMutex.Lock()
	defer dbMutex.Unlock()

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
		SemesterID string `json:"semester_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	res, err := db.Exec("INSERT INTO classes (name, semester_id) VALUES (?, ?)", data.Name, data.SemesterID)
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
		rows.Scan(&class.ID, &class.Name, &class.SemesterID)
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

	dbMutex.Lock()
	defer dbMutex.Unlock()

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
		rows.Scan(&class.ID, &class.Name, &class.SemesterID)
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

	dbMutex.Lock()
	defer dbMutex.Unlock()

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
		SortOrder string `json:"sort_order"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	_, username := isLoggedIn(r)
	user, err := userFromUsername(username)
	if err != nil {
		http.Error(w, "Internal server error - failed to get user", http.StatusInternalServerError)
		return
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	res, err := db.Exec("INSERT INTO semesters (name, sort_order, user_id) VALUES (?, ?, ?)", data.Name, data.SortOrder, user.ID)
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

	if rows.Next() {
		rows.Scan(&semester.ID, &semester.Name, &semester.SortOrder, &semester.UserID)
	} else {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	rows.Close()

	semesters, err := normalizeSemesterSortOrders(w, user.ID)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(semesters)
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

	_, username := isLoggedIn(r)
	user, err := userFromUsername(username)
	if err != nil {
		http.Error(w, "Internal server error - failed to get user", http.StatusInternalServerError)
		return
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

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

	if rows.Next() {
		rows.Scan(&semester.ID, &semester.Name, &semester.SortOrder, &semester.UserID)
	} else {
		http.Error(w, "Internal server error - failed to query database", http.StatusInternalServerError)
		return
	}
	rows.Close()

	semesters, err := normalizeSemesterSortOrders(w, user.ID)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(semesters)
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

	_, username := isLoggedIn(r)
	user, err := userFromUsername(username)
	if err != nil {
		http.Error(w, "Internal server error - failed to get user", http.StatusInternalServerError)
		return
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err = db.Exec("DELETE FROM semesters WHERE id = ?", data.ID)

	if err != nil {
		http.Error(w, "Internal server error - failed to delete semester", http.StatusInternalServerError)
		return
	}

	semesters, err := normalizeSemesterSortOrders(w, user.ID)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(semesters)
}
