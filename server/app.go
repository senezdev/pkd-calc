package server

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type CalcRequest struct {
	Rooms    []string `json:"rooms"`
	TimeLeft string   `json:"time_left"`
	Lobby    string   `json:"lobby"`
}

type CalcResponse struct {
	Error string `json:"error,omitempty"`
}

func calcHandler(w http.ResponseWriter, r *http.Request) {
	var req CalcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var resp CalcResponse

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("Panic: %v\n", err)
				log.Errorf("Stack trace: %s\n", debug.Stack())
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func StartServer() {
	r := mux.NewRouter()

	r.Use(RecoveryMiddleware)
	r.Use(LoggingMiddleware)

	r.HandleFunc("/api/chattriggers/calc", calcHandler).Methods("POST")

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Infof("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
