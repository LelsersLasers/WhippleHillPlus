package main

import (
	"html/template"
	"net/http"
)

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

	classes, assignments := allClassesAndAssignments(user.ID)
	data := map[string]interface{}{
		"user":        user,
		"classes":     classes,
		"assignments": assignments,
	}
	templ := template.Must(template.ParseFiles("templates/home.tmpl"))
	templ.Execute(w, data)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	failContext := map[string]string{
		"email":         "",
		"password":      "",
		"error_message": "",
	}
	failContext = cookiesToFailContext(failContext, &w, r)

	templ := template.Must(template.ParseFiles("templates/login.tmpl"))
	templ.Execute(w, failContext)
}

func registerPage(w http.ResponseWriter, r *http.Request) {
	failContext := map[string]string{
		"email":         "",
		"name":          "",
		"password_1":    "",
		"password_2":    "",
		"error_message": "",
	}
	failContext = cookiesToFailContext(failContext, &w, r)

	templ := template.Must(template.ParseFiles("templates/register.tmpl"))
	templ.Execute(w, failContext)
}
