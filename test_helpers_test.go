package main

// CHQ: Claude AI refactored this file
// test_helpers.go

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)


func TestMain(m *testing.M) {
	fmt.Println("=== TestMain starting ===")

	connStr := os.Getenv("NEON_DB_BSS")
	// fmt.Println("=== connStr:", connStr, "===")
	
	if connStr == "" {
		log.Fatal("NEON_DB_BSS not set")
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to open test db: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping test db: %v", err)
	}

	os.Exit(m.Run())
}

func decodeJSON(t *testing.T, r io.Reader, v any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(v); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}

func muxSetVar(r *http.Request, key string, value int) *http.Request {
	vars := map[string]string{
		key: strconv.Itoa(value),
	}
	return mux.SetURLVars(r, vars)
}

func muxSetVars(r *http.Request, vars map[string]string) *http.Request {
    return mux.SetURLVars(r, vars)
}

func cleanupTestData(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		db.Exec("DELETE FROM gameplay_checkpoints")
		db.Exec("DELETE FROM players")
		db.Exec("DELETE FROM users")
	})
}

func insertUser(t *testing.T, descopeID string) int {
	t.Helper()

	var id int
	err := db.QueryRow(`
		INSERT INTO users (descope_id)
		VALUES ($1)
		RETURNING id
	`, descopeID).Scan(&id)

	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	return id
}

func insertUsername(t *testing.T, userID int, username string) {
	t.Helper()

	_, err := db.Exec(`
		UPDATE users SET username = $1 WHERE id = $2
	`, username, userID)

	if err != nil {
		t.Fatalf("insert username: %v", err)
	}
}

func fetchUsername(t *testing.T, userID int) string {
	t.Helper()

	var username string
	err := db.QueryRow(`
		SELECT username FROM users WHERE id = $1
	`, userID).Scan(&username)

	if err == sql.ErrNoRows {
		return ""
	}

	if err != nil {
		t.Fatalf("fetch username: %v", err)
	}

	return username
}

func insertPlayer(t *testing.T, userID int, playername string) int {
	t.Helper()

	var id int
	err := db.QueryRow(`
		INSERT INTO players (user_id, playername)
		VALUES ($1, $2)
		RETURNING id
	`, userID, playername).Scan(&id)

	if err != nil {
		t.Fatalf("insert player: %v", err)
	}

	return id
}

func insertCheckpoint(t *testing.T, playerID int, title string, data string) int {
	t.Helper()

	var id int
	err := db.QueryRow(`
		INSERT INTO gameplay_checkpoints (player_id, title, data)
		VALUES ($1, $2, $3)
		RETURNING id
	`, playerID, title, data).Scan(&id)

	if err != nil {
		t.Fatalf("insert checkpoint: %v", err)
	}

	return id
}