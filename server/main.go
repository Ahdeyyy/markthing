package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

const secret_key string = "supersecurekeystoredhere"

func main() {
	// Database connection parameters
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	db, err := newDb(dbHost, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer db.Close()

	// Test the database connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	log.Println("Successfully connected to the database")

	handler := newHandler(db)

	// Set up HTTP server
	http.HandleFunc("GET  /users", handler.FindAllUsers)
	http.HandleFunc("POST /user/create", handler.CreateUser)
	http.HandleFunc("POST /user/login", handler.Login)
	http.Handle("GET /protected", handler.AuthMiddleware(handler.ProtectRoute))
	// http.HandleFunc("/", index)
	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func index(w http.ResponseWriter, r *http.Request) {

	log.Println("logging from index route")
	fmt.Fprintf(w, `Hello, World! %s`, r.URL.Path)
}

// func getUsers(w http.ResponseWriter, r *http.Request) {
// 	// Query the database
// 	rows, err := db.Query("SELECT id, name FROM users")
// 	if err != nil {
// 		log.Printf("Error querying database: %v", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}
// 	defer rows.Close()
//
// 	// Process the results
// 	var users []User
// 	for rows.Next() {
// 		var user User
// 		err := rows.Scan(&user.Id, &user.Username)
// 		if err != nil {
// 			log.Printf("Error scanning row: %v", err)
// 			continue
// 		}
// 		users = append(users, user)
// 	}
//
// 	// Check for errors from iterating over rows
// 	if err := rows.Err(); err != nil {
// 		log.Printf("Error iterating over rows: %v", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Send the response
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(users)
// }
