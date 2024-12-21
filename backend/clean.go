package main

import (
	"fmt"
	"time"
)

func deleteInvalidSessions() {
	mutex.Lock()
	defer mutex.Unlock()

	now := time.Now().Unix()
	_, err := db.Exec("DELETE FROM sessions WHERE expiration < ?", now)
	if err != nil {
		fmt.Println("Error deleting sessions: ", err)
	}
}

func startSessionCleanup() {
	ticker := time.NewTicker(CleanInterval)
	defer ticker.Stop()

	for range ticker.C {
		deleteInvalidSessions()
	}
}
