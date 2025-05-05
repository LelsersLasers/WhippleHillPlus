package main

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	PasswordHash string `json:"password_hash"`
	ICSLink      string `json:"ics_link"`
	Timezone     string `json:"timezone"`
}

type Session struct {
	ID         int    `json:"id"`
	Token      string `json:"token"`
	Expiration string `json:"expiration"`
	UserID     int    `json:"user_id"`
}

type Semester struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	UserID    int    `json:"user_id"`
}

type Class struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	SemesterID int    `json:"semester_id"`
}

type Assignment struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	DueDate      string `json:"due_date"`
	DueTime      string `json:"due_time"`
	AssignedDate string `json:"assigned_date"`
	Status       string `json:"status"`
	Type         string `json:"type"`
	ClassID      int    `json:"class_id"`
}
