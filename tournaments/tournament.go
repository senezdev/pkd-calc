package tournaments

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// TODO: don't know which fields need to be exported yet

type participant struct {
	id          uint64
	inGameNick  string
	discordNick string
}

type tournament struct {
	id              uuid.UUID
	startTime       uint64
	longevity       uint64
	participantList []participant
}

func registerTournament(participantList []participant) {
	_, err := uuid.NewV7()
	if err != nil {
		log.Warnf("failed to generate uuid for a tournament: %v", err)
	}

	// t := tournament{
	// 	id:              tournamentId,
	// 	startTime:       startTime,
	// 	longevity:       0,
	// 	participantList: participantList,
	// }

	// TODO: some fancy code that inserts that into DB
}

func RegisterTournamentFromCsv(fileContent []byte) error {
	buf := bytes.NewBuffer(fileContent)
	reader := bufio.NewReader(buf)
	csvReader := csv.NewReader(reader)

	records, err := csvReader.ReadAll()
	if err != nil {
		log.Warn(err)
		return err
	}

	if len(records) == 0 {
		err := fmt.Errorf("This file is empty")
		log.Warn(err)
		return err
	}

	if len(records) == 1 {
		err := fmt.Errorf("This file only has a header (or a single row, which isn't much better)")
		log.Warn(err)
		return err
	}

	var participantList []participant

	for i := 1; i < len(records); i++ {
		id, err := uuid.NewV7()
		if err != nil {
			err = fmt.Errorf("failed generating UUID: %v", err)
			log.Warn(err)
			return err
		}

		if len(records[i]) < 2 {
			err = fmt.Errorf("row %d has less than 2 columns", i)
			log.Warn(err)
			return err
		}
		p := participant{
			id:          uint64(id.ID()),
			inGameNick:  records[i][0],
			discordNick: records[i][1],
		}
		participantList = append(participantList, p)
	}

	registerTournament(participantList)
	log.Infof("Succesfully registered these players: %+v", participantList)
	return nil
}
