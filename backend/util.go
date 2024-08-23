package main

import (
	"fmt"
	"net/http"
	"time"
)

func userFromEmail(email string) (User, error) {
	user := User{}

	mutex.Lock()
	defer mutex.Unlock()

	rows, err := db.Query("SELECT * FROM users WHERE email = ?", email)
	if err != nil {
		return user, err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash)
		return user, nil
	}

	return user, fmt.Errorf("User not found")
}

func isLoggedIn(r *http.Request) (bool, string) {
	cookie, err := r.Cookie(SessionIdCookieName)
	if err != nil {
		return false, ""
	}

	mutex.Lock()
	defer mutex.Unlock()

	rows, err := db.Query("SELECT * FROM users WHERE email = ?", cookie.Value)
	if err != nil {
		return false, cookie.Value
	}
	defer rows.Close()

	return rows.Next(), cookie.Value
}

func login(w *http.ResponseWriter, r *http.Request, email string) {
	// Already verified that login info is correct
	sessionID := email
	http.SetCookie(*w, &http.Cookie{
		Name:    SessionIdCookieName,
		Value:   sessionID,
		Expires: time.Now().Add(SessionTimeout),
		Path:    "/",
	})
	http.Redirect(*w, r, "/", http.StatusSeeOther)
}

func logout(w *http.ResponseWriter, r *http.Request) {
	http.SetCookie(*w, &http.Cookie{
		Name:    SessionIdCookieName,
		Value:   "",
		Expires: time.Now(),
	})

	http.Redirect(*w, r, "/login", http.StatusSeeOther)
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

func allClassesAndAssignments(user_id int) ([]Class, []Assignment) {
	classes := []Class{}
	assignments := []Assignment{}

	mutex.Lock()
	defer mutex.Unlock()

	rows, err := db.Query("SELECT * FROM classes WHERE user_id = ?", user_id)
	if err != nil {
		return classes, assignments
	}
	defer rows.Close()

	for rows.Next() {
		class := Class{}
		rows.Scan(&class.ID, &class.Name, &class.UserID)
		classes = append(classes, class)
	}

	for _, class := range classes {
		rows, err := db.Query("SELECT * FROM assignments WHERE class_id = ?", class.ID)
		if err != nil {
			return classes, assignments
		}
		defer rows.Close()

		for rows.Next() {
			assignment := Assignment{}
			rows.Scan(&assignment.ID, &assignment.Name, &assignment.Description, &assignment.DueDate, &assignment.DueTime, &assignment.AssignedDate, &assignment.Status, &assignment.Type, &assignment.ClassID)
			assignments = append(assignments, assignment)
		}
	}

	return classes, assignments
}
