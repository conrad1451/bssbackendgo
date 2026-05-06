package main

// CHQ: Claude AI refactored this file
// main.go

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time" // Import time package for the timestamp fields

	// PostgreSQL driver
	"github.com/descope/go-sdk/descope/client"
	_ "github.com/lib/pq"

	// Import the handlers package for CORS middleware
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// User represents a user record in the database.
type User struct {
    ID        int       `json:"id"`
    DescopeID string    `json:"descope_id"`
    Username  *string   `json:"username"`
    CreatedAt time.Time `json:"created_at"`
}

type Player struct {
    ID         int       `json:"id"`
    UserID     int       `json:"user_id"`
    Playername *string   `json:"playername"`
    CreatedAt  time.Time `json:"created_at"`
}

// Checkpoint represents a checkpoint in the database.
type Checkpoint struct {
    ID        int       `json:"id"`
    PlayerID  int       `json:"player_id"`
    Title     string    `json:"title"`
    Data      string    `json:"data"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
 
type SuccessResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
}
 
var descopeClient *client.DescopeClient

var listOfDBConnections = []string{"GOOGLE_CLOUD_SQL_BSS", "AVIEN_MYSQL_DB_CONNECTION", "AVIEN_PSQL_DB_CONNECTION", "DIG_OCEAN_DROPLET_PSQL_BSS", "XATA_DB_BSS", "NEON_DB_BSS"}

// mustGetEnv retrieves the value of the required environment variable named by key.
//
// If the environment variable is not set or is empty, the function logs a fatal
// error and terminates the program. This function is intended for mandatory
// configuration values that the application cannot run without, such as
// database connection strings or API credentials.
func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("FATAL: required environment variable %s is not set", key)
	}
	return val
}


// writeJSONError writes a standardized JSON error response.
//
// The function sets the "Content-Type" header to "application/json", writes the
// provided HTTP status code, and encodes a response body containing a success
// flag set to false and an optional error message.
//
// This helper is intended for consistent error responses across HTTP handlers.
// JSON encoding errors are intentionally ignored, as error responses should
// not fail due to secondary encoding issues.
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(SuccessResponse{
		Success: false,
		Message: msg,
	})
}



// writeJSONResponse writes a JSON-encoded HTTP response with the given status code.
//
// The function sets the "Content-Type" header to "application/json", writes the
// provided HTTP status code, and encodes the given payload as JSON in the response body.
//
// The payload may be any value supported by json.Encoder (structs, maps, slices, etc.).
// Encoding errors are intentionally ignored, as this helper is intended for simple,
// best-effort response writing in HTTP handlers.
func writeJSONResponse(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}



func main() { 
	connStr := mustGetEnv(listOfDBConnections[5])

	var err error
	db, err = sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	} 

	if err = db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
  	// defer db.Close()

	// err = db.Ping()
	// if err != nil {
	// 	log.Fatalf("Error connecting to the database: %v", err)
	// }
	fmt.Println("Successfully connected to the database!")

	projectID := os.Getenv("DESCOPE_PROJECT_ID")
	if projectID == "" {
		log.Fatal("DESCOPE_PROJECT_ID environment variable not set.")
	}
	descopeClient, err = client.NewWithConfig(&client.Config{ProjectID: projectID})
	if err != nil {
		log.Fatalf("failed to initialize Descope client: %v", err)
	}

	theOrigins := []string{
		"https://studentfrontendreact-git-test-point-conrad1451s-projects.vercel.app",
		"https://studentfrontendreact.vercel.app",
		"http://localhost:5173",
		"http://localhost:5174",
		"https://*.descope.com",
		"https://static.descope.com", 
	}

	// --- CORS Setup ---
	// Create a list of allowed origins (e.g., your front-end URL)
	allowedOrigins := handlers.AllowedOrigins(theOrigins)

	// Create a list of allowed methods (GET, POST, etc.)
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})

	// Create a list of allowed headers, including Content-Type
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})

	router := mux.NewRouter()
	registerRoutes(router)
	// Wrap your router with the CORS handler
	corsRouter := handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(router)
	// --- End of CORS Setup ---

	// Start the HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	fmt.Printf("Server listening on port %s...\n", port)

	// Pass the corsRouter to ListenAndServe
	log.Fatal(http.ListenAndServe(":"+port, corsRouter))
}
