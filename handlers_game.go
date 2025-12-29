package main // Must match the package name in main.go
import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Checkpoint represents a checkpoint in the database.
// type Checkpoint struct {
//     ID        int       `json:"checkpoint_id"`
//     UserID    string    `json:"user_id"`
//     Title     string    `json:"title"`
//     Data      string    `json:"data"` // Use []byte for JSONB
// 	// Data      []byte    `json:"data"` // Use []byte for JSONB
//     CreatedAt time.Time `json:"created_at"`
//     UpdatedAt time.Time `json:"updated_at"`
// }


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