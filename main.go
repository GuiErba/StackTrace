package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	log.Println("StackTrace server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
