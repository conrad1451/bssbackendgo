package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time" // Import time package for the timestamp fields

	// PostgreSQL driver
	"github.com/descope/go-sdk/descope/client"
	_ "github.com/lib/pq"

	// Import the handlers package for CORS middleware
	"github.com/gorilla/handlers"
)

// Checkpoint represents a checkpoint in the database.
type Checkpoint struct {
    ID        int       `json:"checkpoint_id"`
    UserID    string    `json:"user_id"`
    Title     string    `json:"title"`
    Data      string    `json:"data"` // Use []byte for JSONB
	// Data      []byte    `json:"data"` // Use []byte for JSONB
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type UpdatePlayerRequest struct {
    // FirstName string `json:"first_name"`
    // LastName  string `json:"last_name"`
	Username string `json:"user_name"`
    Email     string `json:"email"`
}

type SuccessResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
}

// User represents a user record in the database.
type User struct {
	UserID   int    `json:"user_id"`
	Username string `json:"user_name"`
}

var db *sql.DB
var descopeClient *client.DescopeClient

var listOfDBConnections = []string{"GOOGLE_CLOUD_SQL_BSS", "AVIEN_MYSQL_DB_CONNECTION", "AVIEN_PSQL_DB_CONNECTION", "DIG_OCEAN_DROPLET_PSQL_BSS", "IBM_DOCKER_PSQL_BSS", "XATA_DB_BSS"}

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
	// Initialize database connection
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

	// // Initialize the router
	// router := mux.NewRouter()

	// // All routes now go through the mux router, including static files
	// router.HandleFunc("/", helloHandler)
	// router.HandleFunc("/favicon.ico", faviconHandler)

	// router.HandleFunc("/health", healthHandler).Methods("GET")


	// // Protected routes (require session validation)
	// protectedRoutes := router.PathPrefix("/api").Subrouter()
	// protectedRoutes.Use(sessionValidationMiddleware) // Apply middleware to all routes in this subrouter
	// // protectedRoutes.HandleFunc("/gamecheckpoints", createCheckpoint).Methods("POST")
	// protectedRoutes.HandleFunc("/gamecheckpoints", getAllBSSCheckpoints).Methods("GET")
	// protectedRoutes.HandleFunc("/gamecheckpoints/{checkpoint_id}", getCheckpoint).Methods("GET")

	// // CHQ: no longer needed since /username already guarntees correctness
	// // - Authenticated user
	// // - Normalized username
	// // - DB unique constraint
	// // - Proper HTTP status codes
	// // protectedRoutes.HandleFunc("/check-username", checkUsername).Methods("GET")
	// protectedRoutes.HandleFunc("/username", setUsername).Methods("POST")
	// protectedRoutes.HandleFunc("/me", getMe).Methods("GET")


	// admin := protectedRoutes.PathPrefix("/admin").Subrouter()
	// admin.Use(requireAdminMiddleware)

	// admin.HandleFunc("/checkpoints", getAllCheckpointsAsAdmin).Methods("GET")

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



// // CHQ: Gemini AI created
// func checkUsername(w http.ResponseWriter, r *http.Request) {
// 	username := strings.TrimSpace(r.URL.Query().Get("username"))

// 	if username == "" {
// 		writeJSONError(w, http.StatusBadRequest, "username is required")
// 		return
// 	}

// 	// Optional: enforce formatting rules here
// 	if len(username) < 3 || len(username) > 20 {
// 		writeJSONResponse(w, http.StatusOK, map[string]bool{
// 			"available": false,
// 		})
// 		return
// 	}

// 	var exists bool
// 	err := db.QueryRow(`
// 		SELECT EXISTS (
// 			SELECT 1
// 			FROM players
// 			WHERE LOWER(user_name) = LOWER($1)
// 		)
// 	`, username).Scan(&exists)

// 	if err != nil {
// 		log.Printf("DB error checking username: %v", err)
// 		writeJSONError(w, http.StatusInternalServerError, "internal error")
// 		return
// 	}

// 	writeJSONResponse(w, http.StatusOK, map[string]bool{
// 		"available": !exists,
// 	})
// }


// CHQ: Gemini AI debugged this function
func getCheckpointOld(w http.ResponseWriter, r *http.Request) {
	// isAdmin, _ := r.Context().Value(contextKeyIsAdmin).(bool)
	
	playerID, ok := r.Context().Value(contextKeyPlayerID).(string)
	if !ok || playerID == "" {
		writeJSONError(w, http.StatusForbidden, "Forbidden: player ID not found in session") 
        return		 
	}

	// declared and not used: idcompilerUnusedVar
	// vars := mux.Vars(r)
	// checkpoint_id, err := strconv.Atoi(vars["checkpoint_id"])
	// if err != nil {
	// 	http.Error(w, "Invalid player ID", http.StatusBadRequest)
	// 	return
	// }

	// CHQ: Gemini AI debugged the function call
	thoseCheckpoints, err := GetUserCheckpoints(playerID)
    if err != nil {
        // If an error occurred in the database function, handle it here.
		log.Printf("DB error retrieving checkpoints: %v",  err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error") 
		return
    }

	// CHQ: Gemini AI debugged the error handling
	// Check if the user has any checkpoints and send an empty array if not.
    if thoseCheckpoints == nil {
        thoseCheckpoints = []Checkpoint{}
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(thoseCheckpoints)
}

// CHQ: ChatGPT generated
func generateUsername(base string, userID int) string {
	base = strings.ToLower(strings.TrimSpace(base))
	if base == "" {
		base = "player"
	}
	return fmt.Sprintf("%s-%06d", base, userID%1_000_000)
}

// CHQ: Gemini AI renamed from getCheckpoints to updateCheckpoint
// func updateCheckpointALT(w http.ResponseWriter, r *http.Request) {
//  vars := mux.Vars(r)
//  id, err := strconv.Atoi(vars["checkpoint_id"])
//  if err != nil {
//      http.Error(w, "Invalid myCheckpoint ID", http.StatusBadRequest)
//      return
//  }

//  var myCheckpoint Checkpoint
//  err = json.NewDecoder(r.Body).Decode(&myCheckpoint)
//  if err != nil {
//      http.Error(w, err.Error(), http.StatusBadRequest)
//      return
//  }

//  if myCheckpoint.ID != 0 && myCheckpoint.ID != checkpoint_id {
//      http.Error(w, "ID in URL and request body do not match", http.StatusBadRequest)
//      return
//  }
//  myCheckpoint.ID = checkpoint_id
//     // Database automatically updates last_edited_at columns
//  query := `UPDATE gameplay_checkpoints SET user_name = $1, checkpoint_data = $2 WHERE checkpoint_id = $3`
//  result, err := db.Exec(query, myCheckpoint.Username, myCheckpoint.Data, myCheckpoint.ID)
//  if err != nil {
//      http.Error(w, fmt.Sprintf("Error updating myCheckpoint: %v", err), http.StatusInternalServerError)
//      return
//  }

//  rowsAffected, err := result.RowsAffected()
//  if err != nil {
//      http.Error(w, fmt.Sprintf("Error checking rows affected: %v", err), http.StatusInternalServerError)
//      return
//  }
//  if rowsAffected == 0 {
//      http.Error(w, "Checkpoint not found or no changes made", http.StatusNotFound)
//      return
//  }

//  w.Header().Set("Content-Type", "application/json")
//  json.NewEncoder(w).Encode(map[string]string{"message": "Checkpoint updated successfully"})
// }
