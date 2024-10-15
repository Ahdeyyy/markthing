package main

import (
	"context"
	"fmt"
	"log"
	"markthing/handler"
	"markthing/store"
	"net/http"
	"os"
)

const secret_key string = "supersecurekeystoredhere"

func main() {
	// Database connection parameters
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	conn, err := store.NewConn(dbHost, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer conn.Close(context.Background())

	// Test the database connection
	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	log.Println("Successfully connected to the database")

	// Set up HTTP server
	params := handler.HandlerParams{Database: conn}
	http.HandleFunc("GET  /users", handler.GetAllUsers(params))
	http.HandleFunc("POST /user/create", handler.CreateUser(params))
	http.HandleFunc("POST /user/login", handler.Login(params))
	http.Handle("GET /protected", handler.AuthMiddleware(params, handler.ProtectedRoute))
	// http.HandleFunc("/", index)
	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	log.Println("logging from index route")
	fmt.Fprintf(w, `Hello, World! %s`, r.URL.Path)
}
