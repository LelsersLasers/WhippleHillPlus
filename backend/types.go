package main

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	PasswordHash string `json:"password_hash"`
}

type Class struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	UserID int    `json:"user_id"`
}


type Assignment struct {
	ID           int    `json:"id"`
	Name         string `json:"title"`
	Description  string `json:"description"`
	DueDate      string `json:"due_date"`
	DueTime      string `json:"due_time"`
	AssignedDate string `json:"assigned_date"`
	Status       string `json:"status"`
	Type         string `json:"type"`
	ClassID      int    `json:"class_id"`
}
