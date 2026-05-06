package main

// CHQ: Claude AI refactored this file
// handlers_user.go

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/lib/pq"
)

// getMe returns the authenticated user's identity and their players.
func getMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(contextKeyUserID).(int)
	if !ok || userID == 0 {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var username *string
	err := db.QueryRowContext(ctx, `
		SELECT username FROM users WHERE id = $1
	`, userID).Scan(&username)

	if err != nil {
		log.Printf("getMe DB error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	players, err := getPlayersByUser(ctx, userID)
	if err != nil {
		log.Printf("getMe players error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]any{
		"user_id":  userID,
		"username": username,
		"players":  players,
	})
}

// setUsername sets or updates the authenticated user's username.
func setUsername(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(contextKeyUserID).(int)
	if !ok || userID == 0 {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	username := strings.TrimSpace(body.Username)
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "username required")
		return
	}

	_, err := db.ExecContext(ctx, `
		UPDATE users SET username = $1 WHERE id = $2
	`, username, userID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			writeJSONError(w, http.StatusConflict, "username already taken")
			return
		}
		log.Printf("setUsername DB error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]bool{"success": true})
}

// createPlayerHandler creates a new player/character for the authenticated user.
func createPlayerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(contextKeyUserID).(int)
	if !ok || userID == 0 {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Playername string `json:"playername"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	playername := strings.TrimSpace(body.Playername)
	if playername == "" {
		writeJSONError(w, http.StatusBadRequest, "playername required")
		return
	}

	playerID, err := createPlayer(ctx, userID, playername)
	if err != nil {
		if err.Error() == "player limit reached" {
			writeJSONError(w, http.StatusForbidden, "player limit of 10 reached")
			return
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			writeJSONError(w, http.StatusConflict, "playername already taken")
			return
		}
		log.Printf("createPlayerHandler DB error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSONResponse(w, http.StatusCreated, map[string]any{
		"player_id":  playerID,
		"playername": playername,
	})
}
