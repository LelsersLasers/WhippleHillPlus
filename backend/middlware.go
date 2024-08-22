package main

import (
	"fmt"
	"net/http"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		fmt.Printf("%s - %s %s %s\n", now.Format("2006-01-02 15:04:05"), r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}
