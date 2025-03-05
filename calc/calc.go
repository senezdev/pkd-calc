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

type boostRoom struct {
	Name      string
	Time      float64
	BoostTime float64
	Quality   MoveQuality
}

type room struct {
	Name          string
	BoostlessTime float64
	BoostStrats   []boostRoom
}

var RoomMap = map[string]room{
	"around pillars": {
		Name:          "Around Pillars",
		BoostlessTime: 17.25,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      11.0,
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
		BoostlessTime: 22.25,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      12.25,
				BoostTime: 3.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      18.5,
				BoostTime: 16.0,
				Quality:   BestMove,
			},
		},
	},
	"castle wall": {
		Name:          "Castle Wall",
		BoostlessTime: 16.0,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      11.25,
				BoostTime: 5,
				Quality:   BestMove,
			},
		},
	},
	"tightrope": {
		Name:          "Tightrope",
		BoostlessTime: 27.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      20.0,
				BoostTime: 2.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      17.5,
				BoostTime: 15.5,
				Quality:   BestMove,
			},
		},
	},
	"early 3+1": {
		Name:          "Early 3+1",
		BoostlessTime: 25.25,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      20.5,
				BoostTime: 1.0,
				Quality:   BrilliantMove,
			},
			{
				Name:      "cp 1-2",
				Time:      13.75,
				BoostTime: 11.75,
				Quality:   BestMove,
			},
		},
	},
	"fence squeeze": {
		Name:          "Fence Squeeze",
		BoostlessTime: 19.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      12.25,
				BoostTime: 2.5,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      15.0,
				BoostTime: 13.0,
				Quality:   BestMove,
			},
		},
	},
	"fences": {
		Name:          "Fences",
		BoostlessTime: 13.5,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      9.5,
				BoostTime: 4.0,
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
		BoostlessTime: 15.25,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      10.5,
				BoostTime: 3.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      11.0,
				BoostTime: 8.0,
				Quality:   BestMove,
			},
		},
	},
	"four towers": {
		Name:          "Four Towers",
		BoostlessTime: 23.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1 + riley",
				Time:      14.0,
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
				Name:      "cp 2-3 + riley",
				Time:      21.0,
				BoostTime: 18.5,
				Quality:   GreatMove,
			},
		},
	},
	"ice": {
		Name:          "Ice",
		BoostlessTime: 17.0,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      14.5,
				BoostTime: 0.5,
				Quality:   BrilliantMove,
			},
			{
				Name:      "cp 1-2",
				Time:      10.75,
				BoostTime: 4.5,
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
		BoostlessTime: 22.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      15.5,
				BoostTime: 4.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      13.5,
				BoostTime: 11.0,
				Quality:   BestMove,
			},
		},
	},
	"ladder tower": {
		Name:          "Ladder Tower",
		BoostlessTime: 25.25,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      12.5,
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
		BoostlessTime: 23.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      18.5,
				BoostTime: 2.0,
				Quality:   BestMove,
			},

			{
				Name:      "cp 1-2",
				Time:      16.5,
				BoostTime: 7.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 2-3",
				Time:      21.5,
				BoostTime: 19.0,
				Quality:   BrilliantMove,
			},
		},
	},
	"quartz climb": {
		Name:          "Quartz Climb",
		BoostlessTime: 19.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				BoostTime: 1.5,
				Time:      14.5,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				BoostTime: 11.5,
				Time:      13.5,
				Quality:   BestMove,
			},
		},
	},
	"quartz temple": {
		Name:          "Quartz Temple",
		BoostlessTime: 16.25,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      8,
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
		BoostlessTime: 12.0,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      9.5,
				BoostTime: 2.0,
				Quality:   GreatMove,
			},
			{
				Name:      "cp 1-2",
				Time:      8.75,
				BoostTime: 6,
				Quality:   BestMove,
			},
		},
	},
	"sandpit": {
		// 13.5 15.5
		Name:          "Sandpit",
		BoostlessTime: 34.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      24.5,
				BoostTime: 1.5,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      22.5,
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
		BoostlessTime: 18.75,
		BoostStrats: []boostRoom{
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
		BoostlessTime: 20.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      15.25,
				BoostTime: 1.5,
				Quality:   BestMove,
			},
			{
				Name:      "cp 2-3",
				Time:      16.5,
				BoostTime: 13.5,
				Quality:   BrilliantMove,
			},
		},
	},
	"slime skip": {
		Name:          "Slime Skip",
		BoostlessTime: 15.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      9.25,
				BoostTime: 1.0,
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
		BoostlessTime: 22.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      10.0,
				BoostTime: 2.0,
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
		BoostlessTime: 18.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      9.5,
				BoostTime: 2.0,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      17.0,
				BoostTime: 14.0,
				Quality:   BrilliantMove,
			},
		},
	},
	"triple trapdoors": {
		Name:          "Triple Trapdoors",
		BoostlessTime: 17.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      12.5,
				BoostTime: 3.0,
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
	"underbridge": {
		Name:          "Underbridge",
		BoostlessTime: 23.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      20.0,
				BoostTime: 5.5,
				Quality:   BrilliantMove,
			},
			{
				Name:      "cp 1-2",
				Time:      10.5,
				BoostTime: 8.0,
				Quality:   BestMove,
			},
		},
	},
	"finish room": {
		Name:          "Finish Room",
		BoostlessTime: 4.5,
		BoostStrats: []boostRoom{
			{
				Name:      "lol",
				Time:      3.0,
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

func calcBoostless(roomList []string) float64 {
	time := 0.0
	for _, room := range roomList {
		time += RoomMap[room].BoostlessTime
	}

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

func calcTwoBoost(roomList []string) ([]calcResult, error) {
	if RoomMap[roomList[len(roomList)-1]].Name != "Finish Room" {
		err := fmt.Errorf("last room is supposed to be finish room. this is a programming error")
		log.Warn(err)
		return nil, err
	}

	boostlessTime := calcBoostless(roomList)
	results := make([]calcResult, 0, 81)

	for i := 0; i < 9; i++ {
		for j := i + 1; j < 9; j++ {
			firstBoostRoom := RoomMap[roomList[i]]
			secondBoostRoom := RoomMap[roomList[j]]

			timeBetweenBoosts := 0.0
			for k := i + 1; k < j; k++ {
				timeBetweenBoosts += RoomMap[roomList[k]].BoostlessTime
			}

			for firstBoostStrat := 0; firstBoostStrat < len(firstBoostRoom.BoostStrats); firstBoostStrat++ {
				for secondBoostStrat := 0; secondBoostStrat < len(secondBoostRoom.BoostStrats); secondBoostStrat++ {
					pacelock := max(0, 60-(timeBetweenBoosts+firstBoostRoom.BoostStrats[firstBoostStrat].Time-firstBoostRoom.BoostStrats[firstBoostStrat].BoostTime+secondBoostRoom.BoostStrats[secondBoostStrat].BoostTime))

					boostTime := boostlessTime - (firstBoostRoom.BoostlessTime - firstBoostRoom.BoostStrats[firstBoostStrat].Time) - (secondBoostRoom.BoostlessTime - secondBoostRoom.BoostStrats[secondBoostStrat].Time) + pacelock

					results = append(results, calcResult{
						time: boostTime,
						boostRooms: []CalcResultBoost{
							{
								Ind:      i,
								StratInd: firstBoostStrat,
							},
							{
								Ind:      j,
								StratInd: secondBoostStrat,
								Pacelock: pacelock,
							},
						},
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

func calcThreeBoost(roomList []string) ([]calcResult, error) {
	if RoomMap[roomList[len(roomList)-1]].Name != "Finish Room" {
		err := fmt.Errorf("last room is supposed to be finish room. this is a programming error")
		log.Warn(err)
		return nil, err
	}

	boostlessTime := calcBoostless(roomList)
	results := make([]calcResult, 0, 729)

	for i := 0; i < 9; i++ {
		for j := i + 1; j < 9; j++ {
			for k := j + 1; k < 9; k++ {
				firstBoostRoom := RoomMap[roomList[i]]
				secondBoostRoom := RoomMap[roomList[j]]
				thirdBoostRoom := RoomMap[roomList[k]]

				timeBetweenBoosts12 := 0.0
				for m := i + 1; m < j; m++ {
					timeBetweenBoosts12 += RoomMap[roomList[m]].BoostlessTime
				}
				timeBetweenBoosts23 := 0.0
				for m := j + 1; m < k; m++ {
					timeBetweenBoosts23 += RoomMap[roomList[m]].BoostlessTime
				}

				for firstBoostStrat := 0; firstBoostStrat < len(firstBoostRoom.BoostStrats); firstBoostStrat++ {
					for secondBoostStrat := 0; secondBoostStrat < len(secondBoostRoom.BoostStrats); secondBoostStrat++ {
						for thirdBoostStrat := 0; thirdBoostStrat < len(thirdBoostRoom.BoostStrats); thirdBoostStrat++ {
							pacelock1 := max(0, 60-(timeBetweenBoosts12+firstBoostRoom.BoostStrats[firstBoostStrat].Time-firstBoostRoom.BoostStrats[firstBoostStrat].BoostTime+secondBoostRoom.BoostStrats[secondBoostStrat].BoostTime))

							pacelock2 := max(0, 60-(timeBetweenBoosts23+secondBoostRoom.BoostStrats[secondBoostStrat].Time-secondBoostRoom.BoostStrats[secondBoostStrat].BoostTime+thirdBoostRoom.BoostStrats[thirdBoostStrat].BoostTime))
							boostTime := boostlessTime - (firstBoostRoom.BoostlessTime - firstBoostRoom.BoostStrats[firstBoostStrat].Time) - (secondBoostRoom.BoostlessTime - secondBoostRoom.BoostStrats[secondBoostStrat].Time) - (thirdBoostRoom.BoostlessTime - thirdBoostRoom.BoostStrats[thirdBoostStrat].Time) + pacelock1 + pacelock2

							results = append(results, calcResult{
								time: boostTime,
								boostRooms: []CalcResultBoost{
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
								},
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

func calcSeedInternal(roomList []string) ([]CalcSeedResult, error) {
	boostlessTime := calcBoostless(roomList)
	res := make([]CalcSeedResult, 0, 5)

	twoBoost, err := calcTwoBoost(roomList)
	if err != nil {
		log.Warn(err)
		return nil, err
	}

	if len(twoBoost) == 0 {
		err := fmt.Errorf("two boost calculation returned an empty array")
		log.Warn(err)
		return nil, err
	}

	threeBoost, err := calcThreeBoost(roomList)
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
	return calcSeedInternal(roomList)
}
