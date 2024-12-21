package main

import (
	"fmt"
	"time"
)

func maybeDeleteInvalidSessions() {
	lastCleanMutex.Lock()
	defer lastCleanMutex.Unlock()

	now := time.Now().Unix()
	interval_unix := int64(CleanInterval.Seconds())
	diff := now - lastClean
	if diff > interval_unix {
		deleteInvalidSessions()
		lastClean = now
	}
}

func deleteInvalidSessions() {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	now := time.Now().Unix()
	_, err := db.Exec("DELETE FROM sessions WHERE expiration < ?", now)
	if err != nil {
		fmt.Println("Error deleting sessions: ", err)
	}
}
