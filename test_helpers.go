package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"testing"
)

func decodeJSON(t *testing.T, r io.Reader, v any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(v); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}

func insertPlayer(t *testing.T, externalID string) int {
	t.Helper()

	var id int
	err := db.QueryRow(`
		INSERT INTO players (player_id)
		VALUES ($1)
		RETURNING id
	`, externalID).Scan(&id)

	if err != nil {
		t.Fatalf("insert player: %v", err)
	}

	return id
}

func insertUsername(t *testing.T, playerID int, username string) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO users (player_id, user_name)
		VALUES ($1, $2)
	`, playerID, username)

	if err != nil {
		t.Fatalf("insert username: %v", err)
	}
}

func fetchUsername(t *testing.T, playerID int) string {
	t.Helper()

	var username string
	err := db.QueryRow(`
		SELECT user_name FROM users WHERE player_id = $1
	`, playerID).Scan(&username)

	if err == sql.ErrNoRows {
		return ""
	}

	if err != nil {
		t.Fatalf("fetch username: %v", err)
	}

	return username
}
