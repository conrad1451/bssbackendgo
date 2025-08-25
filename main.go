package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time" // Import time package for the timestamp fields

	// PostgreSQL driver
	"github.com/descope/go-sdk/descope/client"
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
	playerID	   string    `json:"player_id"`
}

var db *sql.DB
var descopeClient *client.DescopeClient

var isAnAdmin bool
// Define a custom key type to avoid collisions
type contextKey string

const contextKeyUserID contextKey = "userID"
const contextKeyPlayerID contextKey = "playerID" // A key for the player ID


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

	projectID := os.Getenv("DESCOPE_PROJECT_BSS_ID")
	if projectID == "" {
		log.Fatal("DESCOPE_PROJECT_ID environment variable not set.")
	}
	descopeClient, err = client.NewWithConfig(&client.Config{ProjectID: projectID})
	if err != nil {
		log.Fatalf("failed to initialize Descope client: %v", err)
	}

	// Initialize the router
	router := mux.NewRouter()

	// All routes now go through the mux router, including static files
	router.HandleFunc("/", helloHandler)
	router.HandleFunc("/favicon.ico", faviconHandler)

	// Protected routes (require session validation)
    protectedRoutes := router.PathPrefix("/api").Subrouter()
    protectedRoutes.Use(sessionValidationMiddleware) // Apply middleware to all routes in this subrouter
	protectedRoutes.HandleFunc("/gamecheckpoints", createCheckpoint).Methods("POST")
	protectedRoutes.HandleFunc("/gamecheckpoints/{id}", getCheckpoint).Methods("GET")
	protectedRoutes.HandleFunc("/gamecheckpoints", getAllCheckpoints).Methods("GET")
 	protectedRoutes.HandleFunc("/gamecheckpoints/{id}", updateCheckpoint).Methods("PUT")
	// protectedRoutes.HandleFunc("/gamecheckpoints/{id}", updateCheckpointALT).Methods("PATCH")
	protectedRoutes.HandleFunc("/gamecheckpoints/{id}", deleteCheckpoint).Methods("DELETE")

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
    favicon, err := os.ReadFile("./static/calculator.ico")
    if err != nil {
        http.NotFound(w, r)
        return
    }

    // Set the Content-Type header
    w.Header().Set("Content-Type", "image/x-icon")
    
    // Write the file content to the response
    w.Write(favicon)
}

// CHQ: Gemini AI created function
// sessionValidationMiddleware is a middleware to validate the Descope session token.
func sessionValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionToken := r.Header.Get("Authorization")
		if sessionToken == "" {
			http.Error(w, "Unauthorized: No session token provided", http.StatusUnauthorized)
			return
		}

		sessionToken = strings.TrimPrefix(sessionToken, "Bearer ")

		ctx := r.Context()
		authorized, token, err := descopeClient.Auth.ValidateSessionWithToken(ctx, sessionToken)
		if err != nil || !authorized {
			log.Printf("Session validation failed: %v", err)
			http.Error(w, "Unauthorized: Invalid session token", http.StatusUnauthorized)
			return
		}
		if descopeClient.Auth.ValidateRoles(context.Background(), token, []string{"Game Admin"}) {
			isAnAdmin = true
		} else {
			isAnAdmin = false
		}

		userID := token.ID
		// userRole := token.GetTenants()
		// userRole := token.GetTenantValue()
		// userRole := token.GetTenants()
		if userID == "" {
			http.Error(w, "Unauthorized: User ID not found in token", http.StatusUnauthorized)
			return
		}
		
		// For this example, we assume the player ID is the same as the user ID.
		// In a real-world app, you would extract this from custom claims in the token.
		playerID := userID

		// Store the user ID and teacher ID in the request's context
		ctxWithUserID := context.WithValue(ctx, contextKeyUserID, userID)
		ctxWithIDs := context.WithValue(ctxWithUserID, contextKeyPlayerID, playerID)
		
		next.ServeHTTP(w, r.WithContext(ctxWithIDs))
	})
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

// createStudent handles POST requests to create a new student record.
func createCheckpointAsPlayer(w http.ResponseWriter, r *http.Request) {
	playerID, ok := r.Context().Value(contextKeyPlayerID).(string)
	if !ok || playerID == "" {
		http.Error(w, "Forbidden: player ID not found in session", http.StatusForbidden)
		return
	}

	var playerCheckpoint Checkpoint 
	err := json.NewDecoder(r.Body).Decode(&playerCheckpoint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Enforce that the playerCheckpoint being created is associated with the authenticated player.
	playerCheckpoint.playerID = playerID

	// 	ID             int       `json:"id"`
	// Username       string    `json:"user_name"`
	// CheckpointData string    `json:"checkpoint_data"`
	// CreatedAt      time.Time `json:"created_at"`
	// LastEditedAt   time.Time `json:"last_edited_at"`
	// playerID	   string    `json:"player_id"`

	query := `INSERT INTO gameplay_checkpoints (user_name, checkpoint_data, player_id) VALUES ($1, $2, $3) RETURNING id`
	err = db.QueryRow(query, playerCheckpoint.Username, playerCheckpoint.CheckpointData, playerCheckpoint.playerID).Scan(&playerCheckpoint.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating checkpoint: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(playerCheckpoint)
}

func createCheckpoint(w http.ResponseWriter, r *http.Request){
	if (isAnAdmin) {
		createCheckpointAsAdmin(w, r)
	} else {
		createCheckpointAsPlayer(w, r)
	}
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
// getStudent handles GET requests to retrieve a single student by ID, but also checks for ownership.
func getCheckpointAsPlayer(w http.ResponseWriter, r *http.Request) {
	playerID, ok := r.Context().Value(contextKeyPlayerID).(string)
	if !ok || playerID == "" {
		http.Error(w, "Forbidden: player ID not found in session", http.StatusForbidden)
		return
	}
	
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	var myCheckpoint Checkpoint
	// Ensure the checkpoint belongs to the authenticated player.
	query := `SELECT id, user_name, checkpoint_data, created_at, last_edited_at, player_id FROM gameplay_checkpoints WHERE id = $1 AND player_id = $2`
	row := db.QueryRow(query, id, playerID)

	err = row.Scan(&myCheckpoint.ID, &myCheckpoint.Username, &myCheckpoint.CheckpointData, &myCheckpoint.CreatedAt, &myCheckpoint.LastEditedAt, &myCheckpoint.playerID) 
	if err == sql.ErrNoRows {
		http.Error(w, "myCheckpoint not found or not owned by this player", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving myCheckpoint: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(myCheckpoint)
}


func getCheckpoint(w http.ResponseWriter, r *http.Request){
	if (isAnAdmin) {
		getCheckpointAsAdmin(w, r)
	} else {
		getCheckpointAsPlayer(w, r)
	}
}


// getAllCheckpointsAsAdmin handles GET requests to retrieve all myCheckpoint records.
func getAllCheckpointsAsAdmin(w http.ResponseWriter) {
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

func getAllCheckpointsAsPlayer(w http.ResponseWriter, r *http.Request) {
	playerID, ok := r.Context().Value(contextKeyPlayerID).(string)
	if !ok || playerID == "" {
		http.Error(w, "Forbidden: player ID not found in session", http.StatusForbidden)
		return
	}

	var gameplayCheckpoints []Checkpoint
	// CHQ: Gemini AI added the two timestamp columns to the SELECT query
	query := `SELECT id, user_name, checkpoint_data, created_at, last_edited_at FROM gameplay_checkpoints WHERE player_id = $1 ORDER BY id`
	rows, err := db.Query(query, playerID)
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


func getAllCheckpoints(w http.ResponseWriter, r *http.Request){
	if (isAnAdmin) {
		getAllCheckpointsAsAdmin(w)
	} else {
		getAllCheckpointsAsPlayer(w, r)
	}
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

func updateCheckpointAsPlayer(w http.ResponseWriter, r *http.Request) {
	playerID, ok := r.Context().Value(contextKeyPlayerID).(string)
	if !ok || playerID == "" {
		http.Error(w, "Forbidden: player ID not found in session", http.StatusForbidden)
		return
	}

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
	query := `UPDATE gameplay_checkpoints SET user_name = $1, checkpoint_data = $2 WHERE id = $3 AND player_id = $4`
	result, err := db.Exec(query, myCheckpoint.Username, myCheckpoint.CheckpointData, myCheckpoint.ID, playerID)
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
	if (isAnAdmin) {
		updateCheckpointAsAdmin(w, r)
	} else {
		updateCheckpointAsPlayer(w, r)
	}
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

func deleteCheckpointAsPlayer(w http.ResponseWriter, r *http.Request) {
	playerID, ok := r.Context().Value(contextKeyPlayerID).(string)
	if !ok || playerID == "" {
		http.Error(w, "Forbidden: player ID not found in session", http.StatusForbidden)
		return
	}
	
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid myCheckpoint ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM gameplay_checkpoints WHERE id = $1 AND player_id = $2`
	result, err := db.Exec(query, id, playerID)
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
	if (isAnAdmin) {
		deleteCheckpointAsAdmin(w, r)
	} else {
		deleteCheckpointAsPlayer(w, r)
	}
}