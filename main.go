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
	PlayerID       sql.NullString `json:"player_id"` // Use sql.NullString for nullable columns
	// playerID	   string    `json:"player_id"`
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

	// All routes now go through the mux router, including static files
	router.HandleFunc("/", helloHandler)
	router.HandleFunc("/favicon.ico", faviconHandler)

	router.HandleFunc("/gamecheckpoints", createCheckpoint).Methods("POST")
	router.HandleFunc("/gamecheckpoints/{id}", getCheckpoint).Methods("GET")
	router.HandleFunc("/gamecheckpoints", getAllCheckpoints).Methods("GET")
 	router.HandleFunc("/gamecheckpoints/{id}", updateCheckpoint).Methods("PUT")
	// router.HandleFunc("/gamecheckpoints/{id}", updateCheckpointALT).Methods("PATCH")
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
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})

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

// CHQ: Gemini AI generated function
// helloHandler is the function that will be executed for requests to the "/" route.
func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "This is the server for the Bee Swarm Simulator (bss) game - at least the version made by Conrad. It's written in Go (aka GoLang).")
}

// faviconHandler serves the favicon.ico file.
func faviconHandler(w http.ResponseWriter, r *http.Request) {
    // Open the favicon file
    favicon, err := os.ReadFile("./static/beehive1.ico")
    if err != nil {
        http.NotFound(w, r)
        return
    }

    // Set the Content-Type header
    w.Header().Set("Content-Type", "image/x-icon")
    
    // Write the file content to the response
    w.Write(favicon)
}

// createStudent handles POST requests to create a new student record.
func createCheckpointAsAdmin(w http.ResponseWriter, r *http.Request) {
 
	var playerCheckpoint Checkpoint
	err := json.NewDecoder(r.Body).Decode(&playerCheckpoint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
 	query := `INSERT INTO gameplay_checkpoints (user_name, checkpoint_data, player_id) VALUES ($1, $2) RETURNING id`
	err = db.QueryRow(query, playerCheckpoint.Username, playerCheckpoint.CheckpointData).Scan(&playerCheckpoint.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating player checkpoint: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(playerCheckpoint)
}
 
func createCheckpoint(w http.ResponseWriter, r *http.Request){
	createCheckpointAsAdmin(w, r);
} 


func getCheckpointAsAdmin(w http.ResponseWriter, r *http.Request) {
 
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

func getCheckpoint(w http.ResponseWriter, r *http.Request){
	getCheckpointAsAdmin(w, r)
}


// getAllCheckpointsAsAdmin handles GET requests to retrieve all myCheckpoint records.
func getAllCheckpointsAsAdmin(w http.ResponseWriter) {
	var gameplayCheckpoints []Checkpoint
	// CHQ: Gemini AI added the two timestamp columns to the SELECT query
	query := `SELECT id, user_name, checkpoint_data, created_at, last_edited_at, player_id FROM gameplay_checkpoints ORDER BY id`
	// query := `SELECT id, user_name, checkpoint_data, created_at, last_edited_at FROM gameplay_checkpoints ORDER BY id`
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving gameplay_checkpoints: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var myCheckpoint Checkpoint
		// CHQ: Gemini AI added the two timestamp fields to the Scan function

		// err := rows.Scan(&myCheckpoint.ID, &myCheckpoint.Username, &myCheckpoint.CheckpointData, &myCheckpoint.CreatedAt, &myCheckpoint.LastEditedAt)
		err := rows.Scan(&myCheckpoint.ID, &myCheckpoint.Username, &myCheckpoint.CheckpointData, &myCheckpoint.CreatedAt, &myCheckpoint.LastEditedAt, &myCheckpoint.PlayerID)
		
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

func getAllCheckpoints(w http.ResponseWriter, r *http.Request){
	getAllCheckpointsAsAdmin(w)
}

// CHQ: Gemini AI renamed from getCheckpoints to updateCheckpoint
func updateCheckpointAsAdmin(w http.ResponseWriter, r *http.Request) {
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


func updateCheckpoint(w http.ResponseWriter, r *http.Request){
	updateCheckpointAsAdmin(w, r)
}


// CHQ: Gemini AI renamed from getCheckpoints to updateCheckpoint
// func updateCheckpointALT(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	id, err := strconv.Atoi(vars["id"])
// 	if err != nil {
// 		http.Error(w, "Invalid myCheckpoint ID", http.StatusBadRequest)
// 		return
// 	}

// 	var myCheckpoint Checkpoint
// 	err = json.NewDecoder(r.Body).Decode(&myCheckpoint)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	if myCheckpoint.ID != 0 && myCheckpoint.ID != id {
// 		http.Error(w, "ID in URL and request body do not match", http.StatusBadRequest)
// 		return
// 	}
// 	myCheckpoint.ID = id
//     // Database automatically updates last_edited_at columns
// 	query := `UPDATE gameplay_checkpoints SET user_name = $1, checkpoint_data = $2 WHERE id = $3`
// 	result, err := db.Exec(query, myCheckpoint.Username, myCheckpoint.CheckpointData, myCheckpoint.ID)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Error updating myCheckpoint: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Error checking rows affected: %v", err), http.StatusInternalServerError)
// 		return
// 	}
// 	if rowsAffected == 0 {
// 		http.Error(w, "Checkpoint not found or no changes made", http.StatusNotFound)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]string{"message": "Checkpoint updated successfully"})
// }

// deleteCheckpointAsAdmin handles DELETE requests to delete a myCheckpoint record by ID.
func deleteCheckpointAsAdmin(w http.ResponseWriter, r *http.Request) {
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

func deleteCheckpoint(w http.ResponseWriter, r *http.Request){
	deleteCheckpointAsAdmin(w, r)
}