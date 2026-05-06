package main

// CHQ: Claude AI generated this file
// handlers_game_test.go

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestGetAllBSSCheckpoints_NoRoleInContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/players/1/checkpoints", nil)
	rr := httptest.NewRecorder()

	getAllBSSCheckpoints(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestGetAllCheckpointsAsPlayer_Empty(t *testing.T) {
	cleanupTestData(t)

	userID := insertUser(t, "google-600000000")
	playerID := insertPlayer(t, userID, "BeeHero")

	// req := httptest.NewRequest(http.MethodGet, "/api/players/"+string(rune(playerID))+"/checkpoints", nil)
	
	req := httptest.NewRequest(http.MethodGet, "/api/players/"+strconv.Itoa(playerID)+"/checkpoints", nil)
	req = muxSetVar(req, "player_id", playerID)

	ctx := context.WithValue(req.Context(), contextKeyIsAdmin, false)
	ctx = context.WithValue(ctx, contextKeyUserID, userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	getAllCheckpointsAsPlayer(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp []Checkpoint
	decodeJSON(t, rr.Body, &resp)

	if len(resp) != 0 {
		t.Fatalf("expected empty checkpoints")
	}
}

func TestGetAllCheckpointsAsPlayer_WithData(t *testing.T) {
	cleanupTestData(t)

	userID := insertUser(t, "google-700000000")
	playerID := insertPlayer(t, userID, "BeeKing")
	insertCheckpoint(t, playerID, "save1", `{"level":5}`)
	insertCheckpoint(t, playerID, "save2", `{"level":10}`)

	req := httptest.NewRequest(http.MethodGet, "/api/players/1/checkpoints", nil)
	req = muxSetVar(req, "player_id", playerID)

	ctx := context.WithValue(req.Context(), contextKeyIsAdmin, false)
	ctx = context.WithValue(ctx, contextKeyUserID, userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	getAllCheckpointsAsPlayer(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp []Checkpoint
	decodeJSON(t, rr.Body, &resp)

	if len(resp) != 2 {
		t.Fatalf("expected 2 checkpoints, got %d", len(resp))
	}
}

func TestGetAllCheckpointsAsPlayer_Forbidden(t *testing.T) {
	cleanupTestData(t)

	userA := insertUser(t, "google-800000000")
	userB := insertUser(t, "google-900000000")
	playerID := insertPlayer(t, userA, "BeeQueen")

	req := httptest.NewRequest(http.MethodGet, "/api/players/1/checkpoints", nil)
	req = muxSetVar(req, "player_id", playerID)

	ctx := context.WithValue(req.Context(), contextKeyIsAdmin, false)
	ctx = context.WithValue(ctx, contextKeyUserID, userB) // wrong user
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	getAllCheckpointsAsPlayer(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestGetCheckpointAsPlayer_NotFound(t *testing.T) {
	cleanupTestData(t)

	userID := insertUser(t, "google-101000000")
	playerID := insertPlayer(t, userID, "BeeMaster")

	req := httptest.NewRequest(http.MethodGet, "/api/players/"+strconv.Itoa(playerID)+"/checkpoints/9999", nil)
	// req = muxSetVar(req, "player_id", playerID)
	// req = muxSetVar(req, "checkpoint_id", 9999)

	req = muxSetVars(req, map[string]string{
		"player_id":     strconv.Itoa(playerID), 
		"checkpoint_id": "9999",
	})

	ctx := context.WithValue(req.Context(), contextKeyIsAdmin, false)
	ctx = context.WithValue(ctx, contextKeyUserID, userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	getCheckpointAsPlayer(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}