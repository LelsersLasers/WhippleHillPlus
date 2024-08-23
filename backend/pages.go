package main

import (
	"html/template"
	"net/http"
)

func loginPage(w http.ResponseWriter, r *http.Request) {
	failContext := map[string]string{
		"email":         "",
		"password":      "",
		"error_message": "",
	}
	failContext = cookiesToFailContext(failContext, &w, r)

	templ, err := template.ParseFiles("templates/base.tmpl", "templates/login.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	templ, err := template.ParseFiles("templates/base.tmpl", "templates/register.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templ.Execute(w, failContext)
}
