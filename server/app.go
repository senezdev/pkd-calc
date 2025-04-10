package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"pkd-bot/calc"
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
	BoostTime     string                       `json:"boost_time,omitempty"`
	BoostRooms    []discord.BoostRoomsResponse `json:"boost_rooms,omitempty"`
	BoostlessTime string                       `json:"boostless_time,omitempty"`
	Error         string                       `json:"error,omitempty"`
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

	res, resBoostRooms, err := discord.ChattriggersHandle(req.Rooms, req.TimeLeft, req.Lobby, req.Ign, debug)
	if err != nil {
		log.Errorf("Error handling ChatTriggers request: %v", err)
		resp.Error = "Failed to process the request"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}
	resp.BoostTime = discord.FormatTime(res.BoostTime)
	resp.BoostlessTime = discord.FormatTime(res.BoostlessTime)
	resp.BoostRooms = resBoostRooms

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Errorf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

type PkdutilsBoostStrat struct {
	Name      string  `json:"name"`
	Time      float64 `json:"time"`
	BoostTime float64 `json:"boost_time"`
}

type PkdutilsSplit struct {
	BoostlessTime float64              `json:"boostless_time"`
	BoostStrats   []PkdutilsBoostStrat `json:"boost_strats"`
}

type PkdutilsRequest struct {
	Rooms  []string                 `json:"rooms"`
	Splits map[string]PkdutilsSplit `json:"splits"`
}

type PkdutilsBody = struct {
	BoostTime     string                       `json:"boost_time"`
	BoostlessTime string                       `json:"boostless_time"`
	BoostRooms    []discord.BoostRoomsResponse `json:"boost_rooms,omitempty"`
}

type PkdutilsResponse struct {
	Best     PkdutilsBody `json:"best"`
	Personal PkdutilsBody `json:"personal"`
	Error    string       `json:"error,omitempty"`
}

func pkdutilsHandler(w http.ResponseWriter, r *http.Request) {
	var req PkdutilsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = fmt.Errorf("Invalid request body: %v", err)
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("received body: %+v", req)

	var resp PkdutilsResponse
	if len(req.Rooms) != 8 {
		resp.Error = "You didn't pass 8 rooms!"
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	splits := make(map[string]calc.Room)
	for key, room := range req.Splits {
		splits[key] = calc.Room{
			Name:          key,
			BoostlessTime: room.BoostlessTime,
		}

		boostStrats := make([]calc.BoostRoom, len(room.BoostStrats))
		for i, strat := range room.BoostStrats {
			boostStrats[i] = calc.BoostRoom{
				Name:      strat.Name,
				Time:      strat.Time,
				BoostTime: strat.BoostTime,
			}
		}

		splits[key] = calc.Room{
			Name:          key,
			BoostlessTime: room.BoostlessTime,
			BoostStrats:   boostStrats,
		}
	}

	res, err := discord.PkdutilsHandle(req.Rooms, splits)
	if err != nil {
		log.Errorf("Error handling ChatTriggers request: %v", err)
		resp.Error = "Failed to process the request"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp.Best.BoostTime = discord.FormatTime(res.Best.Result.BoostTime)
	resp.Best.BoostlessTime = discord.FormatTime(res.Best.Result.BoostlessTime)
	resp.Best.BoostRooms = res.Best.BoostRooms

	resp.Personal.BoostTime = discord.FormatTime(res.Personal.Result.BoostTime)
	if res.Personal.Result.BoostTime >= res.Personal.Result.BoostlessTime {
		resp.Personal.BoostTime = ""
	}
	resp.Personal.BoostlessTime = discord.FormatTime(res.Personal.Result.BoostlessTime)
	resp.Personal.BoostRooms = res.Personal.BoostRooms

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
		log.Infof("%s %s", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func StartServer() error {
	r := mux.NewRouter()

	r.Use(RecoveryMiddleware)
	r.Use(LoggingMiddleware)

	r.HandleFunc("/api/chattriggers/calc", calcHandler).Methods("POST")
	r.HandleFunc("/api/pkdutils/calc", pkdutilsHandler).Methods("POST")

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
