package calc

import (
	"fmt"
	"image/color"
	_ "image/png"
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"
)

type MoveQuality int

const (
	BestMove MoveQuality = iota
	GreatMove
	BrilliantMove
)

var (
	bestMoveColor      = color.RGBA{155, 199, 0, 200}
	greatMoveColor     = color.RGBA{0, 121, 211, 200}
	brilliantMoveColor = color.RGBA{48, 162, 197, 200}
)

type BoostRoom struct {
	Name      string
	Time      float64
	BoostTime float64
	Quality   MoveQuality
}

type Room struct {
	Name          string
	BoostlessTime float64
	BoostStrats   []BoostRoom
}

var RoomMap = map[string]Room{
	"around pillars": {
		Name:          "Around Pillars",
		BoostlessTime: 16.9,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      10.5,
				BoostTime: 1.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      13.0,
				BoostTime: 10.0,
				Quality:   BestMove,
			},
		},
	},
	"blocks": {
		Name:          "Blocks",
		BoostlessTime: 21.3,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      12.0,
				BoostTime: 3.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      16.9,
				BoostTime: 16.1,
				Quality:   BestMove,
			},
		},
	},
	"castle wall": {
		Name:          "Castle Wall",
		BoostlessTime: 15.7,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      9.5,
				BoostTime: 3.0,
				Quality:   BestMove,
			},
		},
	},
	"tightrope": {
		Name:          "Tightrope",
		BoostlessTime: 27.7,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      19.5,
				BoostTime: 2.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      17.4,
				BoostTime: 15.5,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2 + salami",
				Time:      16.5,
				BoostTime: 14.5,
				Quality:   BestMove,
			},
		},
	},
	"early 3+1": {
		Name:          "Early 3+1",
		BoostlessTime: 24.8,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      20.5,
				BoostTime: 1.0,
				Quality:   BrilliantMove,
			},
			{
				Name:      "cp 1-2",
				Time:      13.4,
				BoostTime: 11.75,
				Quality:   BestMove,
			},
		},
	},
	"fence squeeze": {
		Name:          "Fence Squeeze",
		BoostlessTime: 19.8,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      11.5,
				BoostTime: 2.5,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      14.3,
				BoostTime: 13.0,
				Quality:   BestMove,
			},
		},
	},
	"fences": {
		Name:          "Fences",
		BoostlessTime: 13.0,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      9.5,
				BoostTime: 2.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      10.5,
				BoostTime: 8.5,
				Quality:   BestMove,
			},
		},
	},
	"fortress": {
		Name:          "Fortress",
		BoostlessTime: 14.6,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      10.5,
				BoostTime: 3.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      10.4,
				BoostTime: 7.4,
				Quality:   BestMove,
			},
		},
	},
	"four towers": {
		Name:          "Four Towers",
		BoostlessTime: 22.3,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      13.3,
				BoostTime: 1.5,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      21.5,
				BoostTime: 12.5,
				Quality:   BrilliantMove,
			},
			{
				Name:      "cp 2-3",
				Time:      18.0,
				BoostTime: 15.5,
				Quality:   GreatMove,
			},
		},
	},
	"ice": {
		Name:          "Ice",
		BoostlessTime: 16.7,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      14.5,
				BoostTime: 0.5,
				Quality:   BrilliantMove,
			},
			{
				Name:      "cp 1-2",
				Time:      10.1,
				BoostTime: 4.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 2-3",
				Time:      15.0,
				BoostTime: 13.0,
				Quality:   GreatMove,
			},
		},
	},
	"ladder slide": {
		Name:          "Ladder Slide",
		BoostlessTime: 22.3,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      15.5,
				BoostTime: 4.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      13.0,
				BoostTime: 11.0,
				Quality:   BestMove,
			},
		},
	},
	"ladder tower": {
		Name:          "Ladder Tower",
		BoostlessTime: 24.0,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      12.1,
				BoostTime: 1.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      21.0,
				BoostTime: 18.5,
				Quality:   BrilliantMove,
			},
		},
	},
	"overhead 4b": {
		Name:          "Overhead 4b",
		BoostlessTime: 23.2,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      18.0,
				BoostTime: 2.0,
				Quality:   BestMove,
			},

			{
				Name:      "cp 1-2",
				Time:      16.0,
				BoostTime: 7.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 2-3",
				Time:      19.0,
				BoostTime: 14.3,
				Quality:   BrilliantMove,
			},
		},
	},
	"quartz climb": {
		Name:          "Quartz Climb",
		BoostlessTime: 19.0,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				BoostTime: 1.5,
				Time:      13.5,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				BoostTime: 11.0,
				Time:      13.0,
				Quality:   BestMove,
			},
		},
	},
	"quartz temple": {
		Name:          "Quartz Temple",
		BoostlessTime: 16.00,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      8.0,
				BoostTime: 1,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      14.0,
				BoostTime: 10.0,
				Quality:   BestMove,
			},
		},
	},
	"rng skip": {
		Name:          "Rng Skip",
		BoostlessTime: 11.7,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      7.5,
				BoostTime: 2.0,
				Quality:   GreatMove,
			},
			{
				Name:      "cp 1-2",
				Time:      8.1,
				BoostTime: 6,
				Quality:   BestMove,
			},
		},
	},
	"sandpit": {
		// 13.5 15.5
		Name:          "Sandpit",
		BoostlessTime: 33.8,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      23.8,
				BoostTime: 1.5,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      22.8,
				BoostTime: 13.5,
				Quality:   BestMove,
			},
			{
				Name:      "cp 2-3",
				Time:      31.0,
				BoostTime: 29.0,
				Quality:   BrilliantMove,
			},
		},
	},
	"scatter": {
		Name:          "Scatter",
		BoostlessTime: 18.2,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      13.0,
				BoostTime: 3.5,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      12.5,
				BoostTime: 10.0,
				Quality:   BestMove,
			},
		},
	},
	"slime scatter": {
		Name:          "Slime Scatter",
		BoostlessTime: 19.9,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      14.8,
				BoostTime: 1.5,
				Quality:   BestMove,
			},
			{
				Name:      "cp 2-3",
				Time:      15.6,
				BoostTime: 13.5,
				Quality:   BrilliantMove,
			},
		},
	},
	"slime skip": {
		Name:          "Slime Skip",
		BoostlessTime: 15.5,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      7.0,
				BoostTime: 2.9,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      12.5,
				BoostTime: 10.0,
				Quality:   BestMove,
			},
		},
	},
	"tower tightrope": {
		Name:          "Tower Tightrope",
		BoostlessTime: 22.20,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      10.0,
				BoostTime: 1.5,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      20.0,
				BoostTime: 17.5,
				Quality:   BestMove,
			},
		},
	},
	"triple platform": {
		Name:          "Triple Platform",
		BoostlessTime: 18.3,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      9.0,
				BoostTime: 2.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      16.5,
				BoostTime: 14.0,
				Quality:   BrilliantMove,
			},
		},
	},
	"triple trapdoors": {
		Name:          "Triple Trapdoors",
		BoostlessTime: 17.7,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      12.5,
				BoostTime: 3.0,
				Quality:   BestMove,
			},

			{
				Name:      "cp 1-2",
				Time:      11.5,
				BoostTime: 10.0,
				Quality:   BestMove,
			},
		},
	},
	"underbridge": {
		Name:          "Underbridge",
		BoostlessTime: 23.4,
		BoostStrats: []BoostRoom{
			{
				Name:      "cp 0-1",
				Time:      19.5,
				BoostTime: 2.5,
				Quality:   BrilliantMove,
			},
			{
				Name:      "cp 1-2",
				Time:      9.8,
				BoostTime: 8.0,
				Quality:   BestMove,
			},
		},
	},
	"finish room": {
		Name:          "Finish Room",
		BoostlessTime: 4.4,
		BoostStrats: []BoostRoom{
			{
				Name:      "lol",
				Time:      2.9,
				BoostTime: 0.5,
				Quality:   BestMove,
			},
		},
	},
}

func GetRooms() []string {
	res := make([]string, len(RoomMap)-1)
	i := 0
	for _, v := range RoomMap {
		if v.Name == "Finish Room" {
			continue
		}

		res[i] = strings.ToLower(v.Name)
		i++
	}

	return res
}

func calcBoostless(roomList []string, splits map[string]Room) float64 {
	time := 0.0
	for _, room := range roomList {
		time += splits[room].BoostlessTime
	}

	// timesave := calcTimesave(roomList, nil)

	return time
}

type CalcResultBoost struct {
	Ind      int
	StratInd int
	Pacelock float64
}

type calcResult struct {
	time       float64
	boostRooms []CalcResultBoost
}

const (
	roomOneTimesave float64 = 0.3
	early31Timesave float64 = 0.5
	ibTimesave      float64 = 0.1
	ubTimesave      float64 = 0.2
	ftTimesave      float64 = 0.2
)

func calcTimesave(roomList []string, boostStrat []CalcResultBoost, splits map[string]Room) float64 {
	totalTimesave := 0.0

	totalTimesave += roomOneTimesave

	for i := 1; i < len(roomList); i++ {
		prevRoomName := splits[roomList[i-1]].Name
		currentTimesave := 0.0

		if strings.ToLower(prevRoomName) == "early 3+1" {
			for _, boost := range boostStrat {
				if boost.Ind == i-1 && boost.StratInd == 1 {
					currentTimesave += early31Timesave
					break
				}
			}
		}

		if strings.ToLower(prevRoomName) == "four towers" {
			currentTimesave += ftTimesave
		}

		if strings.ToLower(prevRoomName) == "sandpit" || strings.ToLower(prevRoomName) == "castle wall" {
			currentTimesave += ibTimesave
		}

		if strings.ToLower(prevRoomName) == "underbridge" {
			for _, boost := range boostStrat {
				if boost.Ind == i-1 && boost.StratInd == 1 {
					currentTimesave += ubTimesave
					break
				}
			}
		}

		totalTimesave += currentTimesave
	}

	return totalTimesave
}

func calcTwoBoost(roomList []string, splits map[string]Room) ([]calcResult, error) {
	if strings.ToLower(roomList[len(roomList)-1]) != "finish room" {
		err := fmt.Errorf("last room is supposed to be finish room. this is a programming error")
		log.Warn(err)
		return nil, err
	}

	boostlessTime := calcBoostless(roomList, splits)
	results := make([]calcResult, 0, 81)

	for i := 0; i < 9; i++ {
		for j := i + 1; j < 9; j++ {
			firstBoostRoom := splits[roomList[i]]
			secondBoostRoom := splits[roomList[j]]

			timeBetweenBoosts := 0.0
			for k := i + 1; k < j; k++ {
				timeBetweenBoosts += splits[roomList[k]].BoostlessTime
			}

			for firstBoostStrat := 0; firstBoostStrat < len(firstBoostRoom.BoostStrats); firstBoostStrat++ {
				for secondBoostStrat := 0; secondBoostStrat < len(secondBoostRoom.BoostStrats); secondBoostStrat++ {
					pacelock := max(0, 60-(timeBetweenBoosts+firstBoostRoom.BoostStrats[firstBoostStrat].Time-firstBoostRoom.BoostStrats[firstBoostStrat].BoostTime+secondBoostRoom.BoostStrats[secondBoostStrat].BoostTime))

					boostTime := boostlessTime - (firstBoostRoom.BoostlessTime - firstBoostRoom.BoostStrats[firstBoostStrat].Time) - (secondBoostRoom.BoostlessTime - secondBoostRoom.BoostStrats[secondBoostStrat].Time) + pacelock

					boostStrat := []CalcResultBoost{
						{
							Ind:      i,
							StratInd: firstBoostStrat,
						},
						{
							Ind:      j,
							StratInd: secondBoostStrat,
							Pacelock: pacelock,
						},
					}
					timesave := calcTimesave(roomList, boostStrat, splits)

					results = append(results, calcResult{
						time:       boostTime - timesave,
						boostRooms: boostStrat,
					})
				}
			}
		}
	}

	slices.SortFunc(results, func(a, b calcResult) int {
		if a.time < b.time {
			return -1
		}

		return 1
	})

	return results, nil
}

func calcThreeBoost(roomList []string, splits map[string]Room) ([]calcResult, error) {
	if strings.ToLower(roomList[len(roomList)-1]) != "finish room" {
		err := fmt.Errorf("last room is supposed to be finish room. this is a programming error")
		log.Warn(err)
		return nil, err
	}

	boostlessTime := calcBoostless(roomList, splits)
	results := make([]calcResult, 0, 729)

	for i := 0; i < 9; i++ {
		for j := i + 1; j < 9; j++ {
			for k := j + 1; k < 9; k++ {
				firstBoostRoom := splits[roomList[i]]
				secondBoostRoom := splits[roomList[j]]
				thirdBoostRoom := splits[roomList[k]]

				timeBetweenBoosts12 := 0.0
				for m := i + 1; m < j; m++ {
					timeBetweenBoosts12 += splits[roomList[m]].BoostlessTime
				}
				timeBetweenBoosts23 := 0.0
				for m := j + 1; m < k; m++ {
					timeBetweenBoosts23 += splits[roomList[m]].BoostlessTime
				}

				for firstBoostStrat := 0; firstBoostStrat < len(firstBoostRoom.BoostStrats); firstBoostStrat++ {
					for secondBoostStrat := 0; secondBoostStrat < len(secondBoostRoom.BoostStrats); secondBoostStrat++ {
						for thirdBoostStrat := 0; thirdBoostStrat < len(thirdBoostRoom.BoostStrats); thirdBoostStrat++ {
							pacelock1 := max(0, 60-(timeBetweenBoosts12+firstBoostRoom.BoostStrats[firstBoostStrat].Time-firstBoostRoom.BoostStrats[firstBoostStrat].BoostTime+secondBoostRoom.BoostStrats[secondBoostStrat].BoostTime))

							pacelock2 := max(0, 60-(timeBetweenBoosts23+secondBoostRoom.BoostStrats[secondBoostStrat].Time-secondBoostRoom.BoostStrats[secondBoostStrat].BoostTime+thirdBoostRoom.BoostStrats[thirdBoostStrat].BoostTime))
							boostTime := boostlessTime - (firstBoostRoom.BoostlessTime - firstBoostRoom.BoostStrats[firstBoostStrat].Time) - (secondBoostRoom.BoostlessTime - secondBoostRoom.BoostStrats[secondBoostStrat].Time) - (thirdBoostRoom.BoostlessTime - thirdBoostRoom.BoostStrats[thirdBoostStrat].Time) + pacelock1 + pacelock2

							boostStrat := []CalcResultBoost{
								{
									Ind:      i,
									StratInd: firstBoostStrat,
								},
								{
									Ind:      j,
									StratInd: secondBoostStrat,
									Pacelock: pacelock1,
								},
								{
									Ind:      k,
									StratInd: thirdBoostStrat,
									Pacelock: pacelock2,
								},
							}

							timesave := calcTimesave(roomList, boostStrat, splits)

							results = append(results, calcResult{
								time:       boostTime - timesave,
								boostRooms: boostStrat,
							})
						}
					}
				}
			}
		}
	}

	slices.SortFunc(results, func(a, b calcResult) int {
		if a.time < b.time {
			return -1
		}

		return 1
	})

	return results, nil
}

type CalcSeedResult struct {
	BoostlessTime float64
	BoostTime     float64
	BoostRooms    []CalcResultBoost
}

func mergeSortedResults(a, b []calcResult) []calcResult {
	merged := make([]calcResult, 0, len(a)+len(b))
	i, j := 0, 0

	for i < len(a) && j < len(b) {
		if a[i].time <= b[j].time {
			merged = append(merged, a[i])
			i++
		} else {
			merged = append(merged, b[j])
			j++
		}
	}

	// Append remaining elements from either array
	merged = append(merged, a[i:]...)
	merged = append(merged, b[j:]...)

	return merged
}

func calcSeedInternal(roomList []string, splits map[string]Room) ([]CalcSeedResult, error) {
	boostlessTime := calcBoostless(roomList, splits)
	boostlessTime -= calcTimesave(roomList, nil, splits)

	res := make([]CalcSeedResult, 0, 5)

	twoBoost, err := calcTwoBoost(roomList, splits)
	if err != nil {
		log.Warn(err)
		return nil, err
	}

	if len(twoBoost) == 0 {
		err := fmt.Errorf("two boost calculation returned an empty array")
		log.Warn(err)
		return nil, err
	}

	threeBoost, err := calcThreeBoost(roomList, splits)
	if err != nil {
		log.Warn(err)
		return nil, err
	}

	if len(threeBoost) == 0 {
		err := fmt.Errorf("two boost calculation returned an empty array")
		log.Warn(err)
		return nil, err
	}

	for _, r := range mergeSortedResults(twoBoost, threeBoost) {
		res = append(res, CalcSeedResult{
			BoostlessTime: boostlessTime,
			BoostTime:     r.time,
			BoostRooms:    r.boostRooms,
		})
	}

	return res, nil
}

func CalcSeed(roomList []string) ([]CalcSeedResult, error) {
	if roomList[len(roomList)-1] != "finish room" {
		roomList = append(roomList, "finish room")
	}
	return calcSeedInternal(roomList, RoomMap)
}

func CalcSeedCustom(roomList []string, splits map[string]Room) ([]CalcSeedResult, error) {
	if roomList[len(roomList)-1] != "finish room" {
		roomList = append(roomList, "finish room")
	}
	return calcSeedInternal(roomList, splits)
}
