package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time" // Import time package for the timestamp fields

	// PostgreSQL driver
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	// Import the handlers package for CORS middleware
	"github.com/gorilla/handlers"
)

// Checkpoint represents a user record in the database.
// CHQ: Gemini AI added CreatedAt and LastEditedAt to the struct
type Checkpoint struct {
	ID             int       `json:"id"`
	Username       string    `json:"user_name"`
	CheckpointData string    `json:"checkpoint_data"`
	CreatedAt      time.Time `json:"created_at"`
	LastEditedAt   time.Time `json:"last_edited_at"`
}

var db *sql.DB

var listOfDBConnections = []string{"GOOGLE_CLOUD_SQL_BSS", "AVIEN_MYSQL_DB_CONNECTION", "AVIEN_PSQL_DB_CONNECTION", "GOOGLE_VM_HOSTED_SQL"}

func main() {
	// Initialize database connection
	var err error
	dbConnStr := os.Getenv(listOfDBConnections[3])
	if dbConnStr == "" {
		log.Fatal("DATABASE_URL environment variable not set.")
	}

	db, err = sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	fmt.Println("Successfully connected to the database!")

	// Initialize the router
	router := mux.NewRouter()

	// Define API routes
	router.HandleFunc("/gamecheckpoints", createCheckpoint).Methods("POST")
	router.HandleFunc("/gamecheckpoints/{id}", getCheckpoint).Methods("GET")
	router.HandleFunc("/gamecheckpoints", getAllCheckpoints).Methods("GET")
	// CHQ: Gemini AI changed handler function name and route for PUT request
	router.HandleFunc("/gamecheckpoints/{id}", updateCheckpoint).Methods("PUT")
	router.HandleFunc("/gamecheckpoints/{id}", deleteCheckpoint).Methods("DELETE")

	theOrigins := []string{
		"https://studentfrontendreact-git-test-point-conrad1451s-projects.vercel.app",
		"https://studentfrontendreact.vercel.app",
		"http://localhost:5173",
		"http://localhost:5174",
	}

	// --- CORS Setup ---
	// Create a list of allowed origins (e.g., your front-end URL)
	allowedOrigins := handlers.AllowedOrigins(theOrigins)

	// Create a list of allowed methods (GET, POST, etc.)
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	// Create a list of allowed headers, including Content-Type
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})

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

// createCheckpoint handles POST requests to create a new myCheckpoint record.
func createCheckpoint(w http.ResponseWriter, r *http.Request) {
	var myCheckpoint Checkpoint
	err := json.NewDecoder(r.Body).Decode(&myCheckpoint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
    // Database automatically handles ID and timestamp columns
	query := `INSERT INTO gameplay_checkpoints (user_name, checkpoint_data) VALUES ($1, $2) RETURNING id`
	err = db.QueryRow(query, myCheckpoint.Username, myCheckpoint.CheckpointData).Scan(&myCheckpoint.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating myCheckpoint: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(myCheckpoint)
}

// getCheckpoint handles GET requests to retrieve a single myCheckpoint by ID.
func getCheckpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid myCheckpoint ID", http.StatusBadRequest)
		return
	}

	var myCheckpoint Checkpoint
	// CHQ: Gemini AI added the two timestamp columns to the SELECT query
	query := `SELECT id, user_name, checkpoint_data, created_at, last_edited_at FROM gameplay_checkpoints WHERE id = $1`
	row := db.QueryRow(query, id)
    // CHQ: Gemini AI Added the two timestamp fields to the Scan function
	err = row.Scan(&myCheckpoint.ID, &myCheckpoint.Username, &myCheckpoint.CheckpointData, &myCheckpoint.CreatedAt, &myCheckpoint.LastEditedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "Checkpoint not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving myCheckpoint: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(myCheckpoint)
}

// getAllCheckpoints handles GET requests to retrieve all myCheckpoint records.
func getAllCheckpoints(w http.ResponseWriter, r *http.Request) {
	var gameplayCheckpoints []Checkpoint
	// CHQ: Gemini AI added the two timestamp columns to the SELECT query
	query := `SELECT id, user_name, checkpoint_data, created_at, last_edited_at FROM gameplay_checkpoints ORDER BY id`
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving gameplay_checkpoints: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var myCheckpoint Checkpoint
		// CHQ: Gemini AI added the two timestamp fields to the Scan function
		err := rows.Scan(&myCheckpoint.ID, &myCheckpoint.Username, &myCheckpoint.CheckpointData, &myCheckpoint.CreatedAt, &myCheckpoint.LastEditedAt)
		if err != nil {
			log.Printf("Error scanning myCheckpoint row: %v", err)
			continue
		}
		gameplayCheckpoints = append(gameplayCheckpoints, myCheckpoint)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Error iterating over myCheckpoint rows: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameplayCheckpoints)
}

// CHQ: Gemini AI renamed from getCheckpoints to updateCheckpoint
func updateCheckpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid myCheckpoint ID", http.StatusBadRequest)
		return
	}

	var myCheckpoint Checkpoint
	err = json.NewDecoder(r.Body).Decode(&myCheckpoint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if myCheckpoint.ID != 0 && myCheckpoint.ID != id {
		http.Error(w, "ID in URL and request body do not match", http.StatusBadRequest)
		return
	}
	myCheckpoint.ID = id
    // Database automatically updates last_edited_at columns
	query := `UPDATE gameplay_checkpoints SET user_name = $1, checkpoint_data = $2 WHERE id = $3`
	result, err := db.Exec(query, myCheckpoint.Username, myCheckpoint.CheckpointData, myCheckpoint.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating myCheckpoint: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking rows affected: %v", err), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Checkpoint not found or no changes made", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Checkpoint updated successfully"})
}

// deleteCheckpoint handles DELETE requests to delete a myCheckpoint record by ID.
func deleteCheckpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid myCheckpoint ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM gameplay_checkpoints WHERE id = $1`
	result, err := db.Exec(query, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting myCheckpoint: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking rows affected: %v", err), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Checkpoint not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Checkpoint deleted successfully"})
}