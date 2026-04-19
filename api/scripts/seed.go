package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	godotenv.Load()

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var id, apiKey string
	err = db.QueryRow(`
		INSERT INTO projects (name, owner_email)
		VALUES ('Meu App Teste', 'gui@test.com')
		RETURNING id, api_key
	`).Scan(&id, &apiKey)

	if err != nil {
		log.Fatal("Error inserting project: ", err)
	}

	fmt.Println("=== Projeto de teste criado ===")
	fmt.Printf("Project ID: %s\n", id)
	fmt.Printf("API Key:    %s\n", apiKey)
	fmt.Println("===============================")
	fmt.Printf("\nUse este header para testar:\n")
	fmt.Printf("  -H \"Authorization: Bearer %s\"\n", apiKey)
}
