package main // Must match the package name in main.go
import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// CHQ: Gemini AI generated function
// helloHandler is the function that will be executed for requests to the "/" route.
func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "This is the server for the Bee Swarm Simulator (bss) game - at least the version made by Conrad. It's written in Go (aka GoLang).")
}

// faviconHandler serves the favicon.ico file.
func faviconHandler(w http.ResponseWriter, r *http.Request) {
	// Open the favicon file
	favicon, err := os.ReadFile("./static/beehive1.ico")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", "image/x-icon")

	// Write the file content to the response
	w.Write(favicon)
}
 
// CHQ: ChatGPT created
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

	protected.HandleFunc("/me", getMe).Methods("GET")
	protected.HandleFunc("/username", setUsername).Methods("POST")

	protected.HandleFunc("/gamecheckpoints", getAllBSSCheckpoints).Methods("GET")
	protected.HandleFunc("/gamecheckpoints/{checkpoint_id}", getCheckpoint).Methods("GET")

	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(requireAdminMiddleware)
	admin.HandleFunc("/checkpoints", getAllCheckpointsAsAdmin).Methods("GET")

	
	// --- Disabled until editor UI is ready ---
	// protectedRoutes.HandleFunc("/gamecheckpoints", getAllCheckpoints).Methods("GET")
	// protectedRoutes.HandleFunc("/gamecheckpoints/{checkpoint_id}", updateCheckpoint).Methods("PUT")
	// protectedRoutes.HandleFunc("/gamecheckpoints/{checkpoint_id}", updateCheckpointALT).Methods("PATCH")
	// protectedRoutes.HandleFunc("/gamecheckpoints/{checkpoint_id}", deleteCheckpoint).Methods("DELETE")
}

