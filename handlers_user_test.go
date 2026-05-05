package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMe_NoUserYet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)

	ctx := context.WithValue(req.Context(), contextKeyExternalPlayerID, "descope|test123")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	getMe(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp struct {
		ID       string  `json:"id"`
		Username *string `json:"username"`
	}

	decodeJSON(t, rr.Body, &resp)

	if resp.ID != "descope|test123" {
		t.Fatalf("unexpected id: %s", resp.ID)
	}

	if resp.Username != nil {
		t.Fatalf("expected username to be nil")
	}
}

func TestSetUsername_Success(t *testing.T) {
	externalID := "descope|user123"

	playerID := insertPlayer(t, externalID)

	body := map[string]string{
		"username": "coolname",
	}

	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/username", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), contextKeyExternalPlayerID, externalID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	setUsername(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	username := fetchUsername(t, playerID)
	if username != "coolname" {
		t.Fatalf("expected username to be saved")
	}
}

func TestSetUsername_Conflict(t *testing.T) {
	externalID1 := "descope|userA"
	externalID2 := "descope|userB"

	playerA := insertPlayer(t, externalID1)
	insertUsername(t, playerA, "takenname")

	insertPlayer(t, externalID2)

	body := map[string]string{"username": "takenname"}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/username", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), contextKeyExternalPlayerID, externalID2)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	setUsername(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}
