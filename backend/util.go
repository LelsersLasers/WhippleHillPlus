package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/arran4/golang-ical"
)

func userFromUsername(username string) (User, error) {
	// Don't need to worry about mutex, the calling function will handle it

	user := User{}

	rows, err := db.Query("SELECT * FROM users WHERE username = ?", username)
	if err != nil {
		return user, err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&user.ID, &user.Username, &user.Name, &user.PasswordHash, &user.ICSLink)
		return user, nil
	}

	return user, fmt.Errorf("User not found")
}

func isLoggedIn(r *http.Request) (bool, string) {
	maybeDeleteInvalidSessions()

	cookie, err := r.Cookie(SessionUsernameCookieName)
	if err != nil {
		return false, ""
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	username := cookie.Value

	user, err := userFromUsername(username)
	if err != nil {
		return false, ""
	}

	cookie, err = r.Cookie(SessionTokenCookieName)
	if err != nil {
		return false, username
	}
	token := cookie.Value

	now := time.Now().Unix()
	rows, err := db.Query("SELECT * FROM sessions WHERE token = ? AND expiration > ? AND user_id = ?", token, now, user.ID)
	if err != nil {
		return false, username
	}
	defer rows.Close()

	if rows.Next() {
		return true, username
	} else {
		_, err := db.Exec("DELETE FROM sessions WHERE token = ?", token)
		if err != nil {
			fmt.Println("Error deleting session: ", err)
		}
		return false, username
	}
}

func login(w *http.ResponseWriter, r *http.Request, username string) {
	// Already verified that login info is correct
	// Don't need to worry about mutex, the calling function will handle it

	// Check for existing session
	cookie, err := r.Cookie(SessionTokenCookieName)
	if err == nil {
		_, err := db.Exec("DELETE FROM sessions WHERE token = ?", cookie.Value)
		if err != nil {
			fmt.Println("Error deleting session: ", err)
			return
		}
	}

	// Create new session
	token := uuid.New().String()
	http.SetCookie(*w, &http.Cookie{
		Name:    SessionTokenCookieName,
		Value:   token,
		Expires: time.Now().Add(SessionTimeout),
		Path:    "/",
	})
	http.SetCookie(*w, &http.Cookie{
		Name:    SessionUsernameCookieName,
		Value:   username,
		Expires: time.Now().Add(SessionTimeout),
		Path:    "/",
	})

	user, err := userFromUsername(username)
	if err != nil {
		fmt.Println("Error getting user from username: ", err)
		return
	}

	timestamp := time.Now().Add(SessionTimeout).Unix()
	_, err = db.Exec("INSERT INTO sessions (token, expiration, user_id) VALUES (?, ?, ?)", token, timestamp, user.ID)
	if err != nil {
		fmt.Println("Error inserting session into database: ", err)
	}

	http.Redirect(*w, r, "/", http.StatusSeeOther)
}

func logout(w *http.ResponseWriter, r *http.Request, redirect bool) {
	// Check for existing session
	cookie, err := r.Cookie(SessionTokenCookieName)
	if err == nil {
		dbMutex.Lock()
		defer dbMutex.Unlock()

		_, err := db.Exec("DELETE FROM sessions WHERE token = ?", cookie.Value)
		if err != nil {
			fmt.Println("Error deleting session: ", err)
		}
	}

	http.SetCookie(*w, &http.Cookie{
		Name:    SessionTokenCookieName,
		Value:   "",
		Expires: time.Now(),
	})
	http.SetCookie(*w, &http.Cookie{
		Name:    SessionUsernameCookieName,
		Value:   "",
		Expires: time.Now(),
	})

	if redirect {
		http.Redirect(*w, r, "/login", http.StatusSeeOther)
	}
}

func normalizeSemesterSortOrders(w http.ResponseWriter, user_id int) ([]Semester, error) {
	// Don't need to worry about mutex, the calling function will handle it
	rows, err := db.Query("SELECT * FROM semesters WHERE user_id = ? ORDER BY sort_order", user_id)
	if err != nil {
		http.Error(w, "Internal server error - failed to get semesters", http.StatusInternalServerError)
		return nil, err
	}

	semesters := []Semester{}

	for rows.Next() {
		sem := Semester{}
		rows.Scan(&sem.ID, &sem.Name, &sem.SortOrder, &sem.UserID)
		semesters = append(semesters, sem)
	}
	rows.Close()

	for i, sem := range semesters {
		target_sort_order := i + 1
		if sem.SortOrder != target_sort_order {
			_, err := db.Exec("UPDATE semesters SET sort_order = ? WHERE id = ?", target_sort_order, sem.ID)
			if err != nil {
				http.Error(w, "Internal server error - failed to update semester sort order", http.StatusInternalServerError)
				return nil, err
			}
			semesters[i].SortOrder = target_sort_order
		}
	}

	return semesters, nil
}

func failContextToCookies(w *http.ResponseWriter, failContext map[string]string) {
	for key, value := range failContext {
		http.SetCookie(*w, &http.Cookie{
			Name:    ContextFailCookieNameBase + key,
			Value:   value,
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
				Name:    ContextFailCookieNameBase + key,
				Value:   "",
				Expires: time.Now(),
			})
		}
	}

	return failContext
}

func allSemestersClassesAndAssignments(user_id int) ([]Semester, []Class, []Assignment) {
	semesters := []Semester{}
	classes := []Class{}
	assignments := []Assignment{}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	rows, err := db.Query("SELECT * FROM semesters WHERE user_id = ?", user_id)
	if err != nil {
		return semesters, classes, assignments
	}
	defer rows.Close()

	for rows.Next() {
		semester := Semester{}
		rows.Scan(&semester.ID, &semester.Name, &semester.SortOrder, &semester.UserID)
		semesters = append(semesters, semester)
	}

	for _, semester := range semesters {
		rows, err := db.Query("SELECT * FROM classes WHERE semester_id = ?", semester.ID)
		if err != nil {
			return semesters, classes, assignments
		}
		defer rows.Close()

		for rows.Next() {
			class := Class{}
			rows.Scan(&class.ID, &class.Name, &class.SemesterID)
			classes = append(classes, class)
		}
	}

	for _, class := range classes {
		rows, err := db.Query("SELECT * FROM assignments WHERE class_id = ?", class.ID)
		if err != nil {
			return semesters, classes, assignments
		}
		defer rows.Close()

		for rows.Next() {
			assignment := Assignment{}
			rows.Scan(&assignment.ID, &assignment.Name, &assignment.Description, &assignment.DueDate, &assignment.DueTime, &assignment.AssignedDate, &assignment.Status, &assignment.Type, &assignment.ClassID)
			assignments = append(assignments, assignment)
		}
	}

	return semesters, classes, assignments
}

func getUserIDFromUUID(uuid string) (int, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	rows, err := db.Query("SELECT id FROM users WHERE ics_link = ?", uuid)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	if rows.Next() {
		var userID int
		rows.Scan(&userID)
		return userID, nil
	}

	return -1, fmt.Errorf("User not found")
}

func generateICS(classes []Class, assignments []Assignment) string {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)
	cal.SetName("Assignments Calendar")

	for _, a := range assignments {
		dueTime := a.DueTime
		if dueTime == "" {
			dueTime = ICSDefaultDueTime
		}

		startTime, err := time.Parse("2006-01-02 15:04", a.DueDate+" "+dueTime)
		if err != nil {
			continue
		}

		endTime := startTime.Add(10 * time.Minute)

		class := Class{}
		for _, c := range classes {
			if c.ID == a.ClassID {
				class = c
				break
			}
		}

		classStr := fmt.Sprintf("%s: ", class.Name)

		event := cal.AddEvent(fmt.Sprintf("assignment-%d", a.ID))
		event.SetSummary(fmt.Sprintf("%s%s", classStr, a.Name))
		if a.Description != "" {
			event.SetDescription(a.Description)
		}
		event.SetStartAt(startTime)
		event.SetEndAt(endTime)
		event.SetCreatedTime(time.Now())
		event.SetURL("https://lelserslasers.alwaysdata.net/")
	}

	return cal.Serialize()
}