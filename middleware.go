package main // Must match the package name in main.go
import (
	"context"
	"log"
	"net/http"
	"strings"
)

// Define a custom key type to avoid collisions
type contextKey string

const contextKeyIsAdmin contextKey = "isAdmin"
const contextKeyPlayerID contextKey = "playerID" // int (DB)
const contextKeyExternalPlayerID contextKey = "externalPlayerID" // string (Descope)
const contextKeyUserID contextKey = "userID" // int (DBs)

const contextKeyUsername contextKey = "username"


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
