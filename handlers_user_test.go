package main

// CHQ: Claude AI refactored this file
// handlers_user_test.go

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMe_NoUserYet(t *testing.T) {
	cleanupTestData(t)

	userID := insertUser(t, "google-000000000")

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	ctx := context.WithValue(req.Context(), contextKeyUserID, userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	getMe(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp struct {
		UserID   int      `json:"user_id"`
		Username *string  `json:"username"`
		Players  []Player `json:"players"`
	}
	decodeJSON(t, rr.Body, &resp)

	if resp.UserID != userID {
		t.Fatalf("unexpected user_id: %d", resp.UserID)
	}
	if resp.Username != nil {
		t.Fatalf("expected username to be nil")
	}
	if len(resp.Players) != 0 {
		t.Fatalf("expected no players")
	}
}

func TestGetMe_WithUsername(t *testing.T) {
	cleanupTestData(t)

	userID := insertUser(t, "google-111111111")
	insertUsername(t, userID, "myusername")

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	ctx := context.WithValue(req.Context(), contextKeyUserID, userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	getMe(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp struct {
		UserID   int      `json:"user_id"`
		Username *string  `json:"username"`
		Players  []Player `json:"players"`
	}
	decodeJSON(t, rr.Body, &resp)

	if resp.Username == nil || *resp.Username != "myusername" {
		t.Fatalf("expected username myusername, got %v", resp.Username)
	}
}

func TestGetMe_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rr := httptest.NewRecorder()

	getMe(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestSetUsername_Success(t *testing.T) {
	cleanupTestData(t)

	userID := insertUser(t, "google-222222222")

	body := map[string]string{"username": "coolname"}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/username", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), contextKeyUserID, userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	setUsername(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	username := fetchUsername(t, userID)
	if username != "coolname" {
		t.Fatalf("expected username to be saved")
	}
}

func TestSetUsername_Conflict(t *testing.T) {
	cleanupTestData(t)

	userA := insertUser(t, "google-333333333")
	insertUsername(t, userA, "takenname")

	userB := insertUser(t, "google-444444444")

	body := map[string]string{"username": "takenname"}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/username", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), contextKeyUserID, userB)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	setUsername(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestSetUsername_Empty(t *testing.T) {
	cleanupTestData(t)

	userID := insertUser(t, "google-555555555")

	body := map[string]string{"username": ""}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/username", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), contextKeyUserID, userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	setUsername(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestSetUsername_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/username", bytes.NewReader([]byte(`not json`)))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), contextKeyUserID, 1)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	setUsername(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}