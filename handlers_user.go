 
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
// Example response (username not yet set):
//
//	HTTP/1.1 200 OK
//	Content-Type: application/json
//
//	{
//	  "id": "abc123",
//	  "username": null
//	}
//
// Example response (username set):
//
//	HTTP/1.1 200 OK
//	Content-Type: application/json
//
//	{
//	  "id": "abc123",
//	  "username": "player-004219"
//	}
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
