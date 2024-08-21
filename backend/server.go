package main

import (
	"fmt"
	"net/http"
)


const Port = 8080

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	fmt.Printf("Server is running on port %d\n", Port)

	addr := fmt.Sprintf(":%d", Port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("Error starting server: ", err)
	}
}
