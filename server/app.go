package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"pkd-bot/discord"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type CalcRequest struct {
	Ign      string   `json:"ign"`
	Rooms    []string `json:"rooms"`
	TimeLeft string   `json:"time_left"`
	Lobby    string   `json:"lobby"`
}

type CalcResponse struct {
	BoostTime     string `json:"boost_time,omitempty"`
	BoostlessTime string `json:"boostless_time,omitempty"`
	Error         string `json:"error,omitempty"`
}

func calcHandler(w http.ResponseWriter, r *http.Request) {
	var req CalcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = fmt.Errorf("Invalid request body: %v", err)
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("received body: %+v", req)

	debug := r.FormValue("debug") == "true"

	var resp CalcResponse
	if len(req.Rooms) != 8 {
		resp.Error = "You didn't pass 8 rooms!"
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	res, err := discord.ChattriggersHandle(req.Rooms, req.TimeLeft, req.Lobby, req.Ign, debug)
	if err != nil {
		log.Errorf("Error handling ChatTriggers request: %v", err)
		resp.Error = "Failed to process the request"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}
	resp.BoostTime = discord.FormatTime(res.BoostTime)
	resp.BoostlessTime = discord.FormatTime(res.BoostlessTime)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Errorf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
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

func StartServer() error {
	r := mux.NewRouter()

	r.Use(RecoveryMiddleware)
	r.Use(LoggingMiddleware)

	r.HandleFunc("/api/chattriggers/calc", calcHandler).Methods("POST")

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Infof("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Error(err)
		return err
	}

	return nil
}
