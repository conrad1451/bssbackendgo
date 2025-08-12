package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	// PostgreSQL driver
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	// Import the handlers package for CORS middleware
	"github.com/gorilla/handlers"
)

// User represents a user record in the database.
type Checkpoint struct {
	ID        int    `json:"id"`
	Username string `json:"user_name"`
	CheckpointData string `json:"checkpoint_data"`
}



var db *sql.DB
  
var listOfDBConnections = []string{"GOOGLE_CLOUD_SQL_BSS", "AVIEN_MYSQL_DB_CONNECTION", "AVIEN_PSQL_DB_CONNECTION"}

func main() {
	fmt.Println("Please update something!")

	// Initialize database connection
	var err error
	dbConnStr := os.Getenv(listOfDBConnections[0])
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
	router.HandleFunc("/gamecheckpoints/{id}", getCheckpoints).Methods("PUT")
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

	query := `INSERT INTO gamecheckpoints (user_name, checkpoint_data) VALUES ($1, $2) RETURNING id`
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
	query := `SELECT id, user_name, checkpoint_data FROM gamecheckpoints WHERE id = $1`
	row := db.QueryRow(query, id)

	err = row.Scan(&myCheckpoint.ID, &myCheckpoint.Username, &myCheckpoint.CheckpointData)
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
	var gamecheckpoints []Checkpoint
	query := `SELECT id, user_name, checkpoint_data FROM gamecheckpoints ORDER BY id`
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving gamecheckpoints: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var myCheckpoint Checkpoint
		err := rows.Scan(&myCheckpoint.ID, &myCheckpoint.Username, &myCheckpoint.CheckpointData)
		if err != nil {
			log.Printf("Error scanning myCheckpoint row: %v", err)
			continue
		}
		gamecheckpoints = append(gamecheckpoints, myCheckpoint)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Error iterating over myCheckpoint rows: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gamecheckpoints)
}

// getCheckpoints handles PUT requests to update an existing myCheckpoint record.
func getCheckpoints(w http.ResponseWriter, r *http.Request) {
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
 
	query := `UPDATE gamecheckpoints SET user_name = $1, checkpoint_data = $2 WHERE id = $3`
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

	query := `DELETE FROM gamecheckpoints WHERE id = $1`
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