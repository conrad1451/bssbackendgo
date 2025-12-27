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
	"github.com/lib/pq"
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

// Define a custom key type to avoid collisions
type contextKey string

const contextKeyIsAdmin contextKey = "isAdmin"
const contextKeyPlayerID contextKey = "playerID" // int (DB)
const contextKeyExternalPlayerID contextKey = "externalPlayerID" // string (Descope)
const contextKeyUserID contextKey = "userID" // int (DBs)

const contextKeyUsername contextKey = "username"

var listOfDBConnections = []string{"GOOGLE_CLOUD_SQL_BSS", "AVIEN_MYSQL_DB_CONNECTION", "AVIEN_PSQL_DB_CONNECTION", "DIG_OCEAN_DROPLET_PSQL_BSS", "IBM_DOCKER_PSQL_BSS"}

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
	connStr := mustGetEnv(listOfDBConnections[4])

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

	// Initialize the router
	router := mux.NewRouter()

	// All routes now go through the mux router, including static files
	router.HandleFunc("/", helloHandler)
	router.HandleFunc("/favicon.ico", faviconHandler)

	router.HandleFunc("/health", healthHandler).Methods("GET")


	// Protected routes (require session validation)
	protectedRoutes := router.PathPrefix("/api").Subrouter()
	protectedRoutes.Use(sessionValidationMiddleware) // Apply middleware to all routes in this subrouter
	// protectedRoutes.HandleFunc("/gamecheckpoints", createCheckpoint).Methods("POST")
	protectedRoutes.HandleFunc("/gamecheckpoints", getAllBSSCheckpoints).Methods("GET")
	protectedRoutes.HandleFunc("/gamecheckpoints/{checkpoint_id}", getCheckpoint).Methods("GET")

	// CHQ: no longer needed since /username already guarntees correctness
	// - Authenticated user
	// - Normalized username
	// - DB unique constraint
	// - Proper HTTP status codes
	// protectedRoutes.HandleFunc("/check-username", checkUsername).Methods("GET")
	protectedRoutes.HandleFunc("/username", setUsername).Methods("POST")
	protectedRoutes.HandleFunc("/me", getMe).Methods("GET")

	// --- Disabled until editor UI is ready ---
	// protectedRoutes.HandleFunc("/gamecheckpoints", getAllCheckpoints).Methods("GET")
	// protectedRoutes.HandleFunc("/gamecheckpoints/{checkpoint_id}", updateCheckpoint).Methods("PUT")
	// protectedRoutes.HandleFunc("/gamecheckpoints/{checkpoint_id}", updateCheckpointALT).Methods("PATCH")
	// protectedRoutes.HandleFunc("/gamecheckpoints/{checkpoint_id}", deleteCheckpoint).Methods("DELETE")



	admin := protectedRoutes.PathPrefix("/admin").Subrouter()
	admin.Use(requireAdminMiddleware)

	admin.HandleFunc("/checkpoints", getAllCheckpointsAsAdmin).Methods("GET")



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

// setUsername handles requests to set or update the authenticated user's username.
//
// The handler expects a JSON request body of the form:
//
//	{ "username": "desired_name" }
//
// Authentication context must provide an external player identifier via
// contextKeyExternalPlayerID. If no corresponding player record exists, one is
// created automatically.
//
// Behavior:
//   - Validates that the request body contains a non-empty username
//   - Ensures a player record exists for the authenticated user
//   - Inserts or updates the user's username for that player
//   - Enforces username uniqueness at the database level
//
// Responses:
//   - 200 OK on success
//   - 400 Bad Request if the JSON is invalid or the username is empty
//   - 409 Conflict if the username is already taken
//   - 500 Internal Server Error for unexpected database or server errors
func setUsername(w http.ResponseWriter, r *http.Request) {
  externalID := r.Context().Value(contextKeyExternalPlayerID).(string)

  var body struct {
    Username string `json:"username"`
  }

  if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
    writeJSONError(w, 400, "invalid json")
    return
  }

  username := strings.TrimSpace(body.Username)
  if username == "" {
    writeJSONError(w, 400, "username required")
    return
  }

  // ensure player exists
  var playerID int
  err := db.QueryRow(`
    SELECT id FROM players WHERE player_id = $1
  `, externalID).Scan(&playerID)

  if err == sql.ErrNoRows {
    err = db.QueryRow(`
      INSERT INTO players (player_id)
      VALUES ($1)
      RETURNING id
    `, externalID).Scan(&playerID)
  }

  if err != nil {
    writeJSONError(w, 500, "internal error")
    return
  }

  // upsert user + username
  _, err = db.Exec(`
    INSERT INTO users (player_id, user_name)
    VALUES ($1, $2)
    ON CONFLICT (player_id)
    DO UPDATE SET user_name = EXCLUDED.user_name
  `, playerID, username)

  if err != nil {
    if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
      writeJSONError(w, 409, "username already taken")
      return
    }
    writeJSONError(w, 500, "internal error")
    return
  }

  writeJSONResponse(w, 200, map[string]bool{"success": true})
}

 
// CHQ: ChatGPT created
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// resolveOrCreatePlayer resolves the internal player record for an authenticated user.
//
// This function enforces the system’s identity boundary:
//
//   - External identity (Descope) is used ONLY for authentication
//   - Internal identity (players.id) is used for authorization and ownership
//
// Behavior:
//   - If a player row already exists for the given external Descope ID,
//     its internal primary key (players.id) is returned.
//   - If no such row exists, a new player row is created and its internal ID
//     is returned.
//
// Invariants:
//   - The external Descope ID is never used as a database primary key.
//   - The returned value is always an internal integer ID.
//   - This function must be called only after session validation succeeds.
//
// Parameters:
//   - ctx: request-scoped context (must not be nil)
//   - db: database connection
//   - externalPlayerID: Descope-assigned user identifier (string)
//
// Returns:
//   - int: internal player ID (players.id)
//   - error: non-nil if resolution or creation fails
//
// Callers MUST:
//   - Store the returned ID in request context
//   - Use the returned ID for all authorization and ownership checks
// func resolveOrCreatePlayer(ctx context.Context, db *sql.DB, externalPlayerID string) (string, error) {
func resolveOrCreatePlayer(
	ctx context.Context,
	externalPlayerID string,
) (int, error) {

	var playerID int

	err := db.QueryRowContext(ctx, `
		SELECT id FROM players WHERE player_id = $1
	`, externalPlayerID).Scan(&playerID)

	if err == sql.ErrNoRows {
		err = db.QueryRowContext(ctx, `
			INSERT INTO players (player_id)
			VALUES ($1)
			RETURNING id
		`, externalPlayerID).Scan(&playerID)
	}

	if err != nil {
		return 0, err
	}

	return playerID, nil
}



// resolveOrCreateUser resolves the internal user record associated with a player.
//
// This function represents the second layer of identity resolution:
//
//   - players represent authenticated identities (one per Descope user)
//   - users represent application-level entities that own gameplay data
//
// Behavior:
//   - If a user row already exists for the given internal player ID,
//     its primary key (users.id) is returned.
//   - If no such row exists, a new user row is created and its ID is returned.
//
// Invariants:
//   - The input playerID MUST be an internal database ID (players.id).
//   - External identity values (e.g. Descope IDs) must never be passed here.
//   - The returned value is always an internal integer ID.
//   - At most one user row exists per player.
//
// Parameters:
//   - ctx: request-scoped context (must not be nil)
//   - db: database connection
//   - playerID: internal player ID (players.id)
//
// Returns:
//   - int: internal user ID (users.id)
//   - error: non-nil if resolution or creation fails
//
// Callers MUST:
//   - Call this only after resolveOrCreatePlayer succeeds
//   - Store the returned user ID in request context
//   - Use the returned ID for all gameplay ownership checks
func resolveOrCreateUser(
	ctx context.Context,
	db *sql.DB,
	playerID int,
) (int, string, error) {

	var userID int
	var username sql.NullString

	err := db.QueryRowContext(ctx, `
		SELECT user_id, user_name
		FROM users
		WHERE player_id = $1
	`, playerID).Scan(&userID, &username)

	if err == sql.ErrNoRows {
		// create user
		err = db.QueryRowContext(ctx, `
			INSERT INTO users (player_id)
			VALUES ($1)
			RETURNING user_id
		`, playerID).Scan(&userID)
		if err != nil {
			return 0, "", err
		}

		// username intentionally empty
		return userID, "", nil
	}

	if err != nil {
		return 0, "", err
	}

	if username.Valid {
		return userID, username.String, nil
	}

	return userID, "", nil
}

// requireAdminMiddleware wraps an HTTP handler and enforces that the request
// is made by an authenticated admin user.
//
// The middleware expects a boolean admin flag to be present in the request
// context under contextKeyIsAdmin. If the value is missing or false, the
// request is rejected.
//
// Behavior:
//   - Allows the request to proceed if the user is an admin
//   - Returns 403 Forbidden with a JSON error response otherwise
//
// This middleware should be applied after authentication and context
// population middleware that sets contextKeyIsAdmin.
func requireAdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		isAdmin, ok := r.Context().Value(contextKeyIsAdmin).(bool)
		if !ok || !isAdmin {
			writeJSONError(w, http.StatusForbidden, "Forbidden: admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CHQ: Gemini AI created function
// sessionValidationMiddleware validates an incoming request's Descope session
// token and enriches the request context with authenticated user information.
//
// The middleware performs the following steps:
//
//  1. Extracts a Bearer token from the Authorization header
//  2. Validates the session using the Descope Auth API
//  3. Extracts the external Descope player ID from the validated token
//  4. Determines whether the user has the "Game Admin" role
//  5. Stores authentication and authorization data in the request context
//
// On success, the middleware adds the following values to the request context:
//   - contextKeyExternalPlayerID: the Descope player ID (string)
//   - contextKeyIsAdmin: whether the user has the "Game Admin" role (bool)
//   - contextKeyPlayerID: the internal player database ID (int)
//
// If validation fails at any step, the request is terminated with a JSON error
// response and an appropriate HTTP status code:
//
//   - 401 Unauthorized if the session token is missing or invalid
//   - 500 Internal Server Error if user or player resolution fails
//
// This middleware must be executed before any handlers or middleware that rely
// on authenticated user identity or authorization, such as admin-only routes
// or endpoints that require a resolved player.
//
// Note: Authentication must occur before database-backed user resolution.
// This middleware assumes that downstream handlers will use the context values
// rather than re-validating the session token.
func sessionValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 1️⃣ Read Authorization header
		sessionToken := r.Header.Get("Authorization")
		if sessionToken == "" {
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized: No session token provided")
 			return
		}

		sessionToken = strings.TrimPrefix(sessionToken, "Bearer ") 
		ctx := r.Context()

		// 2️⃣ Validate Descope session	
		authorized, token, err := descopeClient.Auth.ValidateSessionWithToken(ctx, sessionToken)
		if err != nil || !authorized {
			log.Printf("Session validation failed: %v", err)
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized: Invalid session token" )
			return
		}
		
		// isAdmin := descopeClient.Auth.ValidateRoles(context.Background(), token, []string{"Game Admin"})
		// isAdmin := descopeClient.Auth.ValidateRoles(ctx, token, []string{"Game Admin"})
		// ctx = context.WithValue(ctx, contextKeyIsAdmin, isAdmin)

		// 3️⃣ Extract external player ID
		descopePlayerID := token.ID
		if descopePlayerID == "" {
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized: Player ID missing")
			return
		}

		isAdmin := descopeClient.Auth.ValidateRoles(ctx, token, []string{"Game Admin"})

		// // ---- 3. Begin transaction ----
		// tx, err := db.BeginTx(ctx, &sql.TxOptions{
		// 	Isolation: sql.LevelReadCommitted,
		// })
		// if err != nil {
		// 	log.Printf("Failed to begin transaction: %v", err)
		// 	writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		// 	return
		// }

		// // Ensure rollback on any failure
		// defer tx.Rollback()
		
		// ---- 4. Resolve / create player ----
		// playerDBID, err := resolveOrCreatePlayer(ctx, db, descopePlayerID)
		// playerDBID, err := resolveOrCreatePlayer(ctx, descopePlayerID)

		if err != nil {
			log.Printf("player resolution failed: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

 
		// ---- 5. Resolve / create user ----
		// 5️⃣ Resolve user (same pattern)
		// userDBID, username, err := resolveOrCreateUser(ctx, db, playerDBID)
		if err != nil {
			log.Printf("user resolution failed: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}


		
		// // ---- 6. Commit transaction ----
		// if err := tx.Commit(); err != nil {
		// 	log.Printf("Transaction commit failed: %v", err)
		// 	writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		// 	return
		// }

		// --- NEW CODE FOR AUTOMATIC REGISTRATION ---
        // This is where a successfully authenticated user is automatically added to the players table.
        // It's called after validation but before processing the request, ensuring the player ID is in the DB.
        // insertPlayerIntoDB(playerID)

		// ---- 7. Store values in request context ---- 
		ctx = context.WithValue(ctx, contextKeyIsAdmin, isAdmin)
		ctx = context.WithValue(ctx, contextKeyExternalPlayerID, descopePlayerID)
		ctx = context.WithValue(ctx, contextKeyPlayerID, playerDBID)
		// ctx = context.WithValue(ctx, contextKeyUserID, userDBID)
		// ctx = context.WithValue(ctx, contextKeyUsername, username)
 
		// ---- 8. Continue request ----
		next.ServeHTTP(w, r.WithContext(ctx)) 
	})
}


// CHQ: Gemini AI refactored function to account for new user table 
//      access in the database
func createCheckpointAsPlayer(w http.ResponseWriter, r *http.Request) {
	playerID, ok := r.Context().Value(contextKeyPlayerID).(int)
	if !ok || playerID == 0 {
		writeJSONError(w, http.StatusForbidden, "Forbidden: player ID not found in session")
		return
	}

	var input struct {
		Title string          `json:"title"`
		Data  json.RawMessage `json:"checkpoint_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if input.Title == "" {
		writeJSONError(w, http.StatusBadRequest, "Title is required")
		return
	}

	query := `
		INSERT INTO gameplay_checkpoints (
			player_id,
			title,
			checkpoint_data,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING
			checkpoint_id,
			player_id,
			title,
			checkpoint_data,
			created_at,
			updated_at
	`

	var cp Checkpoint

	err := db.QueryRow(
		query,
		playerID,
		input.Title,
		input.Data,
	).Scan(
		&cp.ID, 
		&cp.Title,
		&cp.Data,
		&cp.CreatedAt,
		&cp.UpdatedAt,
	)

	if err != nil {
		log.Printf("DB error creating checkpoint: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cp)
}


func createCheckpoint(w http.ResponseWriter, r *http.Request) {
 
	isAdmin, ok := r.Context().Value(contextKeyIsAdmin).(bool)
    if !ok {
        // Fallback for safety, though middleware should ensure it's set
		writeJSONError(w, http.StatusForbidden, "Forbidden: Role not determined") 
        return
    }

	if (isAdmin) {
		// createCheckpointAsAdmin(w, r)
	} else {
		createCheckpointAsPlayer(w, r)
	}
}

// CHQ: Gemini AI refactored to account for fk of user_name and user table
 func getCheckpointAsAdmin(w http.ResponseWriter, r *http.Request) {
 	vars := mux.Vars(r)
	checkpointIDStr := vars["checkpointID"]

	checkpointID, err := strconv.Atoi(checkpointIDStr)
	if err != nil || checkpointID <= 0 {
		writeJSONError(w, http.StatusBadRequest, "Invalid checkpoint ID")
		return
	}

	query := `
		SELECT
			checkpoint_id,
			player_id,
			title,
			checkpoint_data,
			created_at,
			updated_at
		FROM gameplay_checkpoints
		WHERE checkpoint_id = $1
	`

	var cp Checkpoint

	err = db.QueryRow(query, checkpointID).Scan(
		&cp.ID,
		&cp.UserID,
		&cp.Title,
		&cp.Data,
		&cp.CreatedAt,
		&cp.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		writeJSONError(w, http.StatusNotFound, "Checkpoint not found")
		return
	}

	if err != nil {
		log.Printf("DB error retrieving checkpoint %d: %v", checkpointID, err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cp)
}


// getStudent handles GET requests to retrieve a single student by ID, but also checks for ownership.
func getCheckpointAsPlayer(w http.ResponseWriter, r *http.Request) {
	playerID, ok := r.Context().Value(contextKeyPlayerID).(string)
	if !ok || playerID == "" { 
		writeJSONError(w, http.StatusForbidden, "Forbidden: player ID not found in session") 
        return
	}

	vars := mux.Vars(r)
	checkpoint_id, err := strconv.Atoi(vars["checkpoint_id"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid checkpoint ID")
		return
	}

	var myCheckpoint Checkpoint
	var userName string // New variable to hold the user_name from the join
	// Ensure the oldcheckpoint belongs to the authenticated player.
	query := `
		SELECT 
			g.checkpoint_id, 
			u.user_name, 
			g.checkpoint_data, 
			g.created_at, 
			g.last_edited_at, 
			g.player_id 
		FROM 
			gameplay_checkpoints g 
		JOIN 
			users u ON g.user_id = u.user_id 
		WHERE 
			g.checkpoint_id = $1 AND g.player_id = $2` 

	row := db.QueryRow(query, checkpoint_id, playerID)
 	err = row.Scan(
		&myCheckpoint.ID,
		&userName, // Scan into a separate variable
		&myCheckpoint.Data,
		&myCheckpoint.CreatedAt,
		&myCheckpoint.UpdatedAt,
		&myCheckpoint.Title,
	)
	
	if err == sql.ErrNoRows {
		writeJSONError(w, http.StatusNotFound, "myCheckpoint not found or not owned by this player")
		return
	} else if err != nil {
		log.Printf("DB error retrieving checkpoint %d: %v", checkpoint_id, err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(myCheckpoint)
}

func getCheckpoint(w http.ResponseWriter, r *http.Request){	
	// Retrieve isAdmin from context
    isAdmin, ok := r.Context().Value(contextKeyIsAdmin).(bool)
    if !ok {
        // Fallback for safety, though middleware should ensure it's set
		writeJSONError(w, http.StatusForbidden, "Forbidden: player ID not found in session")
        return
    }

	if (isAdmin) {
		getCheckpointAsAdmin(w, r)
	} else {
		getCheckpointAsPlayer(w, r)
	}
}

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

// CHQ: Gemini AI refactored to account for fk of user_name and user table
// getAllCheckpointsAsAdmin handles GET requests to retrieve all myCheckpoint records.
// func getAllCheckpointsAsAdmin(w http.ResponseWriter) {
func getAllCheckpointsAsAdmin(w http.ResponseWriter, r *http.Request) {
	isAdmin, ok := r.Context().Value(contextKeyIsAdmin).(bool)
	if !ok || !isAdmin {
		writeJSONError(w, http.StatusForbidden, "Forbidden: admin access required")
		return
	}

	var checkpoints []Checkpoint

	query := `
		SELECT
			checkpoint_id,
			player_id,
			checkpoint_data,
			created_at,
			last_edited_at,
			title
		FROM gameplay_checkpoints
		ORDER BY checkpoint_id
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("DB error retrieving checkpoints (admin): %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var cp Checkpoint

		err := rows.Scan(
			&cp.ID,
 			&cp.Data,
			&cp.CreatedAt,
			&cp.UpdatedAt,
			&cp.Title,
		)
		if err != nil {
			log.Printf("Error scanning admin checkpoint row: %v", err)
			continue
		}

		checkpoints = append(checkpoints, cp)
	}

	if err := rows.Err(); err != nil {
		log.Printf("DB iteration error (admin): %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(checkpoints)
}


// CHQ: Gemini AI refactored to account for fk of user_name and user table
func getAllCheckpointsAsPlayer(w http.ResponseWriter, r *http.Request) {
	playerID, ok := r.Context().Value(contextKeyPlayerID).(int)
	if !ok {
		writeJSONError(w, http.StatusForbidden, "Forbidden: player not authenticated")
		return
	}

	var gameplayCheckpoints []Checkpoint

	query := `
		SELECT
			checkpoint_id,
			checkpoint_data,
			created_at,
			last_edited_at,
			title
		FROM gameplay_checkpoints
		WHERE player_id = $1
		ORDER BY checkpoint_id
	`

	rows, err := db.Query(query, playerID)
	if err != nil {
		log.Printf("DB error retrieving checkpoints: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var cp Checkpoint

		err := rows.Scan(
			&cp.ID,
			&cp.Data,
			&cp.CreatedAt,
			&cp.UpdatedAt,
			&cp.Title,
		)
		if err != nil {
			log.Printf("Error scanning checkpoint row: %v", err)
			continue
		}

		gameplayCheckpoints = append(gameplayCheckpoints, cp)
	}

	if err := rows.Err(); err != nil {
		log.Printf("DB iteration error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameplayCheckpoints)
}

func getAllBSSCheckpoints(w http.ResponseWriter, r *http.Request){
	// CHQ: Gemini AI changed fetching global vatiable to retrieving variable from context

	// Retrieve isAdmin from context
    isAdmin, ok := r.Context().Value(contextKeyIsAdmin).(bool)
    if !ok {
        // Fallback for safety, though middleware should ensure it's set
		writeJSONError(w, http.StatusForbidden, "Forbidden: Role not determined")
        return
    }

	if (isAdmin) {
		getAllCheckpointsAsAdmin(w, r)
	} else {
		getAllCheckpointsAsPlayer(w, r)
	}
}

// CHQ: ChatGPT generated
func generateUsername(base string, userID int) string {
	base = strings.ToLower(strings.TrimSpace(base))
	if base == "" {
		base = "player"
	}
	return fmt.Sprintf("%s-%06d", base, userID%1_000_000)
}

// getMe returns the authenticated user's identity information.
//
// The handler requires a valid session and expects the authenticated
// external player ID to be present in the request context (populated by
// sessionValidationMiddleware).
//
// The response includes:
//   - id: the external Descope player ID
//   - username: the user's chosen username, or null if none has been set
//
// The username is resolved by joining the players and users tables. If the
// player exists but no associated user or username is found, the username
// field will be null. If the player record does not yet exist, the handler
// also returns username as null.
//
// Error responses:
//   - 401 Unauthorized if the request is missing authentication context
//   - 500 Internal Server Error if a database error occurs
//
// Successful responses always return HTTP 200 with a JSON body containing
// the user's ID and username.
func getMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	externalID, ok := ctx.Value(contextKeyExternalPlayerID).(string)
	if !ok || externalID == "" {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var username *string

	err := db.QueryRow(`
		SELECT u.user_name
		FROM players p
		LEFT JOIN users u ON u.player_id = p.id
		WHERE p.player_id = $1
	`, externalID).Scan(&username)

	if err == sql.ErrNoRows {
		// Player does not exist yet
		writeJSONResponse(w, http.StatusOK, map[string]any{
			"id":       externalID,
			"username": nil,
		})
		return
	}

	if err != nil {
		log.Printf("getMe DB error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{
		"id":       externalID,
		"username": username,
	})
}




func updatePlayerProfile(w http.ResponseWriter, r *http.Request) {
    // 1. Authorization: Get the ID of the logged-in player from the context.
    playerID, ok := r.Context().Value(contextKeyPlayerID).(string)
    if !ok || playerID == "" {
        // This should not happen if middleware succeeded, but good to check.
		writeJSONError(w, http.StatusForbidden, "Forbidden: Player ID not found in session context.")
        return
    }

    // 2. Decode Request Body: CRITICAL FIX
    var req UpdatePlayerRequest // Use the struct with correct JSON tags
    // err := json.NewDecoder(r.Body).Decode(&req)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)

    if err != nil {
 		log.Printf("Invalid request body or JSON format: %v", err)
		writeJSONError(w, http.StatusBadRequest, "Invalid request body or JSON format") 
		return
    }
    
    // 3. Database Update: Update the player profile identified by the authenticated playerID.
    // The query and arguments were correct, but the data source (req) must be correct.
    // CHQ: Gemini AI fixed the line below for correct name for id field
	// FIX: Changed "id" to "player_id" to match the actual database column name.

	// // CHQ: Gemini AI corrected query
	// query := `UPDATE players SET user_name = $1, email = $2 WHERE player_id = $3`    
    // // Note: We are using the fields from the unmarshalled 'req' struct.
    // // Ensure db is available in scope.
    // result, err := db.Exec(query, req.Username, req.Email, playerID)

	query := `
	UPDATE users
	SET user_name = $1
	WHERE player_id = $2`
	_, err = db.Exec(query, req.Username, playerID)


    // if err != nil {
    //     log.Printf("Error executing SQL update for player ID %s: %v", playerID, err)
    //     // Ensure the error response is JSON for the frontend to handle gracefully.
	// 	log.Printf("DB error updating player: %v", err)
	// 	writeJSONError(w, http.StatusInternalServerError, "Internal server error")
	// 	return 
    // }

	// CHQ: ChatGPT wrapped with unique-constraint handling
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			writeJSONError(w, http.StatusConflict, "username already taken")
			return
		}

		log.Printf("DB error updating player: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// CHQ: Gemini AI removed since no longer needed
	// - unique constraint + auth already guarantees correctness.
    // // 4. Check Rows Affected
    // rowsAffected, err := result.RowsAffected()
    // if err != nil {
    //     log.Printf("Error checking rows affected for player ID %s: %v", playerID, err)

	// 	writeJSONError(w, http.StatusInternalServerError, "Error confirming profile update.")
    //     return
    // }
    // if rowsAffected == 0 {
    //     log.Printf("Update attempted for player ID %s, but 0 rows affected. Profile not found?", playerID)
    //     writeJSONError(w, http.StatusNotFound, "Authenticated player profile not found or no changes made")
    //     return
    // }

    // 5. Success Response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
   json.NewEncoder(w).Encode(SuccessResponse{
    Success: true,
    Message: "player updated successfully",
})
}

// CHQ: Gemini AI renamed from getCheckpoints to updateCheckpoint
func updateCheckpointAsAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	checkpoint_id, err := strconv.Atoi(vars["checkpoint_id"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid myCheckpoint ID")
		return
	}

	var myCheckpoint Checkpoint
	// err = json.NewDecoder(r.Body).Decode(&myCheckpoint)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&myCheckpoint)

	if err != nil {
		log.Printf("Invalid request body: %v", err)
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if myCheckpoint.ID != 0 && myCheckpoint.ID != checkpoint_id {
		writeJSONError(w, http.StatusBadRequest, "ID in URL and request body do not match")
		return
	}
	myCheckpoint.ID = checkpoint_id
	// Database automatically updates last_edited_at columns
	query := `UPDATE gameplay_checkpoints SET checkpoint_data = $1 WHERE checkpoint_id = $2`
	result, err := db.Exec(query, myCheckpoint.Data, myCheckpoint.ID)
	if err != nil {
		log.Printf("DB error updating checkpoint: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("DB error checking affected rows: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if rowsAffected == 0 {
		writeJSONError(w, http.StatusNotFound,  "Checkpoint not found or no changes made")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SuccessResponse{
    	Success: true,
    	Message: "Checkpoint updated successfully",
	})

 }

func updateCheckpointAsPlayer(w http.ResponseWriter, r *http.Request) {
	playerID, ok := r.Context().Value(contextKeyPlayerID).(string)
	if !ok || playerID == "" {
		writeJSONError(w, http.StatusForbidden,  "Forbidden: player ID not found in session")
		return
	}

	vars := mux.Vars(r)
	checkpoint_id, err := strconv.Atoi(vars["checkpoint_id"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid myCheckpoint ID")
		return
	}

	var myCheckpoint Checkpoint
	// err = json.NewDecoder(r.Body).Decode(&myCheckpoint)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&myCheckpoint)

	if err != nil {
		log.Printf("Invalid request body: %v", err)
		writeJSONError(w, http.StatusBadRequest, "Invalid request body") 
		return
	}

	if myCheckpoint.ID != 0 && myCheckpoint.ID != checkpoint_id {
		writeJSONError(w, http.StatusBadRequest, "ID in URL and request body do not match")
		return
	}
	myCheckpoint.ID = checkpoint_id
	// Database automatically updates last_edited_at columns
	query := `UPDATE gameplay_checkpoints SET checkpoint_data = $1 WHERE checkpoint_id = $2 AND player_id = $3`
	result, err := db.Exec(query, myCheckpoint.Data, myCheckpoint.ID, playerID)
	if err != nil { 
		log.Printf("DB error updating checkpoint: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error") 
 		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("DB error checking affected rows: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
 		return
	}
	if rowsAffected == 0 {
		writeJSONError(w, http.StatusNotFound, "Checkpoint not found or not owned by this player")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SuccessResponse{
    	Success: true,
    	Message: "Checkpoint updated successfully",
	})

}

func updateCheckpoint(w http.ResponseWriter, r *http.Request) {
	isAdmin, _ := r.Context().Value(contextKeyIsAdmin).(bool)
	// if isAnAdmin {

	if isAdmin {
		updateCheckpointAsAdmin(w, r)
	} else {
		updateCheckpointAsPlayer(w, r)
	}
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

// deleteCheckpointAsAdmin handles DELETE requests to delete a myCheckpoint record by ID.
func deleteCheckpointAsAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	checkpoint_id, err := strconv.Atoi(vars["checkpoint_id"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid myCheckpoint ID")
		return
	}

	query := `DELETE FROM gameplay_checkpoints WHERE checkpoint_id = $1`
	result, err := db.Exec(query, checkpoint_id)
	if err != nil {
		log.Printf("DB error deleting checkpoint for player: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error") 
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("DB error checking rows affected: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if rowsAffected == 0 {
		writeJSONError(w, http.StatusNotFound, "Checkpoint not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SuccessResponse{
    Success: true,
    Message: "Checkpoint deleted successfully",
})
}

func deleteCheckpointAsPlayer(w http.ResponseWriter, r *http.Request) {
	playerID, ok := r.Context().Value(contextKeyPlayerID).(string)
	if !ok || playerID == "" {
		writeJSONError(w, http.StatusForbidden, "Forbidden: player ID not found in session")
	 	return
	}

	vars := mux.Vars(r)
	checkpoint_id, err := strconv.Atoi(vars["checkpoint_id"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid myCheckpoint ID")
		return
	}

	query := `DELETE FROM gameplay_checkpoints WHERE checkpoint_id = $1 AND player_id = $2`
	result, err := db.Exec(query, checkpoint_id, playerID)
	if err != nil {
		log.Printf("DB error deleting checkpoint for player: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("DB error checking rows affected: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
 		return
	}
	if rowsAffected == 0 {
		writeJSONError(w, http.StatusNotFound, "Checkpoint not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	
	json.NewEncoder(w).Encode(SuccessResponse{
    	Success: true,
		Message: "Checkpoint deleted successfully",
	})
}

func deleteCheckpoint(w http.ResponseWriter, r *http.Request) {
	isAdmin, _ := r.Context().Value(contextKeyIsAdmin).(bool)
	// if isAnAdmin {
	if isAdmin {
		deleteCheckpointAsAdmin(w, r)
	} else {
		deleteCheckpointAsPlayer(w, r)
	}
}


// NEW CHECKPOINTS


// CHQ: Gemini AI created function
// Example function to retrieve a user's checkpoints
func GetUserCheckpoints(userID string) ([]Checkpoint, error) {
    // 1. Prepare a slice to hold the checkpoints.
    var checkpoints []Checkpoint

    // 2. Query the database for all checkpoints belonging to the given user ID.
    // The query uses a parameterized statement ($1) to prevent SQL injection.
    rows, err := db.Query("SELECT checkpoint_id, title, data, created_at, updated_at FROM game_checkpoints WHERE user_id = $1", userID)
    if err != nil {
        return nil, fmt.Errorf("failed to query game_checkpoints for user %s: %w", userID, err)
    }
    defer rows.Close()

    // 3. Iterate through the result set and scan each row into a Checkpoint struct.
    for rows.Next() {
        var cp Checkpoint
        err := rows.Scan(&cp.ID, &cp.Title, &cp.Data, &cp.CreatedAt, &cp.UpdatedAt)
        if err != nil {
            return nil, fmt.Errorf("failed to scan checkpoint row: %w", err)
        }
        checkpoints = append(checkpoints, cp)
    }

    // 4. Check for any errors that occurred during the iteration.
    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error during row iteration: %w", err)
    }

    // 5. Return the slice of checkpoints.
    return checkpoints, nil
}