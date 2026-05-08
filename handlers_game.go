package main

// CHQ: Claude AI refactored this file
// main.go

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func getAllBSSCheckpoints(w http.ResponseWriter, r *http.Request) {
	isAdmin, ok := r.Context().Value(contextKeyIsAdmin).(bool)
	if !ok {
		writeJSONError(w, http.StatusForbidden, "role not determined")
		return
	}

	if isAdmin {
		getAllCheckpointsAsAdmin(w, r)
	} else {
		getAllCheckpointsAsPlayer(w, r)
	}
}

func getAllCheckpointsAsPlayer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(contextKeyUserID).(int)
	if !ok || userID == 0 {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	playerID, err := strconv.Atoi(vars["player_id"])
	if err != nil || playerID <= 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid player ID")
		return
	}

	// verify ownership
	var exists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM players WHERE id = $1 AND user_id = $2
		)
	`, playerID, userID).Scan(&exists)
	if err != nil {
		log.Printf("getAllCheckpointsAsPlayer ownership check error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !exists {
		writeJSONError(w, http.StatusForbidden, "forbidden")
		return
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, player_id, title, data, created_at, updated_at
		FROM gameplay_checkpoints
		WHERE player_id = $1
		ORDER BY created_at DESC
	`, playerID)
	if err != nil {
		log.Printf("getAllCheckpointsAsPlayer DB error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer rows.Close()

	var checkpoints []Checkpoint
	for rows.Next() {
		var cp Checkpoint
		if err := rows.Scan(&cp.ID, &cp.PlayerID, &cp.Title, &cp.Data, &cp.CreatedAt, &cp.UpdatedAt); err != nil {
			log.Printf("getAllCheckpointsAsPlayer scan error: %v", err)
			continue
		}
		checkpoints = append(checkpoints, cp)
	}

	if err := rows.Err(); err != nil {
		log.Printf("getAllCheckpointsAsPlayer rows error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if checkpoints == nil {
		checkpoints = []Checkpoint{}
	}

	writeJSONResponse(w, http.StatusOK, checkpoints)
}

func getAllCheckpointsAsAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	isAdmin, ok := ctx.Value(contextKeyIsAdmin).(bool)
	if !ok || !isAdmin {
		writeJSONError(w, http.StatusForbidden, "admin access required")
		return
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, player_id, title, data, created_at, updated_at
		FROM gameplay_checkpoints
		ORDER BY id
	`)
	if err != nil {
		log.Printf("getAllCheckpointsAsAdmin DB error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer rows.Close()

	var checkpoints []Checkpoint
	for rows.Next() {
		var cp Checkpoint
		if err := rows.Scan(&cp.ID, &cp.PlayerID, &cp.Title, &cp.Data, &cp.CreatedAt, &cp.UpdatedAt); err != nil {
			log.Printf("getAllCheckpointsAsAdmin scan error: %v", err)
			continue
		}
		checkpoints = append(checkpoints, cp)
	}

	if err := rows.Err(); err != nil {
		log.Printf("getAllCheckpointsAsAdmin rows error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if checkpoints == nil {
		checkpoints = []Checkpoint{}
	}

	writeJSONResponse(w, http.StatusOK, checkpoints)
}

func getCheckpoint(w http.ResponseWriter, r *http.Request) {
	isAdmin, ok := r.Context().Value(contextKeyIsAdmin).(bool)
	if !ok {
		writeJSONError(w, http.StatusForbidden, "role not determined")
		return
	}

	if isAdmin {
		getCheckpointAsAdmin(w, r)
	} else {
		getCheckpointAsPlayer(w, r)
	}
}

func getCheckpointAsPlayer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(contextKeyUserID).(int)
	if !ok || userID == 0 {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	playerID, err := strconv.Atoi(vars["player_id"])
	if err != nil || playerID <= 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid player ID")
		return
	}

	checkpointID, err := strconv.Atoi(vars["checkpoint_id"])
	if err != nil || checkpointID <= 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid checkpoint ID")
		return
	}

	var cp Checkpoint
	err = db.QueryRowContext(ctx, `
		SELECT id, player_id, title, data, created_at, updated_at
		FROM gameplay_checkpoints
		WHERE id = $1
		AND player_id = $2
		AND EXISTS (
			SELECT 1 FROM players WHERE id = $2 AND user_id = $3
		)
	`, checkpointID, playerID, userID).Scan(
		&cp.ID, &cp.PlayerID, &cp.Title, &cp.Data, &cp.CreatedAt, &cp.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		writeJSONError(w, http.StatusNotFound, "checkpoint not found")
		return
	}
	if err != nil {
		log.Printf("getCheckpointAsPlayer DB error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSONResponse(w, http.StatusOK, cp)
}

func getCheckpointAsAdmin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	checkpointID, err := strconv.Atoi(vars["checkpoint_id"])
	if err != nil || checkpointID <= 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid checkpoint ID")
		return
	}

	var cp Checkpoint
	err = db.QueryRowContext(ctx, `
		SELECT id, player_id, title, data, created_at, updated_at
		FROM gameplay_checkpoints
		WHERE id = $1
	`, checkpointID).Scan(
		&cp.ID, &cp.PlayerID, &cp.Title, &cp.Data, &cp.CreatedAt, &cp.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		writeJSONError(w, http.StatusNotFound, "checkpoint not found")
		return
	}
	if err != nil {
		log.Printf("getCheckpointAsAdmin DB error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSONResponse(w, http.StatusOK, cp)
}

// CHQ: created by Claude AI
func createCheckpointHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(contextKeyUserID).(int)
	if !ok || userID == 0 {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	playerID, err := strconv.Atoi(vars["player_id"])
	if err != nil || playerID <= 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid player ID")
		return
	}

	// verify ownership
	var exists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM players WHERE id = $1 AND user_id = $2
		)
	`, playerID, userID).Scan(&exists)
	if err != nil {
		log.Printf("createCheckpointHandler ownership check error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !exists {
		writeJSONError(w, http.StatusForbidden, "forbidden")
		return
	}

	var body struct {
		Title string          `json:"title"`
		Data  json.RawMessage `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if body.Title == "" {
		writeJSONError(w, http.StatusBadRequest, "title is required")
		return
	}

	var cp Checkpoint
	err = db.QueryRowContext(ctx, `
		INSERT INTO gameplay_checkpoints (player_id, title, data)
		VALUES ($1, $2, $3)
		RETURNING id, player_id, title, data, created_at, updated_at
	`, playerID, body.Title, body.Data).Scan(
		&cp.ID, &cp.PlayerID, &cp.Title, &cp.Data, &cp.CreatedAt, &cp.UpdatedAt,
	)

	if err != nil {
		log.Printf("createCheckpointHandler DB error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSONResponse(w, http.StatusCreated, cp)
}