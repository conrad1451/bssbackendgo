package main

// CHQ: Claude AI refactored this file
// routes.go

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "This is the server for the Bee Swarm Simulator (bss) game - at least the version made by Conrad. It's written in Go (aka GoLang).")
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	favicon, err := os.ReadFile("./static/beehive1.ico")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/x-icon")
	w.Write(favicon)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func registerRoutes(router *mux.Router) {
	router.HandleFunc("/", helloHandler)
	router.HandleFunc("/health", healthHandler)

	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(sessionValidationMiddleware)

	// user
	protected.HandleFunc("/me", getMe).Methods("GET")
	protected.HandleFunc("/username", setUsername).Methods("POST")

	// players
	protected.HandleFunc("/players", createPlayerHandler).Methods("POST")
	protected.HandleFunc("/players/{player_id}/checkpoints", getAllBSSCheckpoints).Methods("GET")
	protected.HandleFunc("/players/{player_id}/checkpoints/{checkpoint_id}", getCheckpoint).Methods("GET")
	protected.HandleFunc("/players/{player_id}/checkpoints", createCheckpointHandler).Methods("POST")

	// admin
	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(requireAdminMiddleware)
	admin.HandleFunc("/checkpoints", getAllCheckpointsAsAdmin).Methods("GET")
}