package hypixel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

var baseURL = "https://api.hypixel.net/v2"

type countGame struct {
	Players int            `json:"players"`
	Modes   map[string]int `json:"modes"`
}

type getCountBody struct {
	Success     bool                 `json:"success"`
	PlayerCount int                  `json:"playerCount"`
	Games       map[string]countGame `json:"games"`
}

func GetPlayerCount() (int, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", baseURL, "counts"), nil)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	req.Header.Add("API-Key", os.Getenv("HYPIXEL_API_KEY"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err)
		return 0, err
	}
	defer resp.Body.Close()

	var body getCountBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		log.Error(err)
		return 0, err
	}

	log.Debugf("%+v", body)

	for k, g := range body.Games["DUELS"].Modes {
		if k == "DUELS_PARKOUR_EIGHT" {
			return g, nil
		}
	}

	err = fmt.Errorf("couldn't find pkd player count")
	log.Error(err)
	return 0, err
}
