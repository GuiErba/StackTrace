package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/joho/godotenv"
	"stacktrace/internal/cache"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	cache.Init()

	err := cache.Client.Set(cache.Ctx, "logpilot:test", "ok", 0).Err()
	if err != nil {
		log.Fatal("Redis connection failed: ", err)
	}
	log.Println("Redis connected!")

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	log.Println("StackTrace server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
