package calc

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"os"

	"github.com/fogleman/gg"
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

var roomMap = map[string]room{
	"Around Pillars": {
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
	"Blocks": {
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
	"Castle Wall": {
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
	"Tightrope": {
		Name:          "Tightrope",
		BoostlessTime: 27.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      20.0,
				BoostTime: 7.5,
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
	"Early 3+1": {
		Name:          "Early 3+1",
		BoostlessTime: 25.25,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      21.0,
				BoostTime: 8.0,
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
	"Fence Squeeze": {
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
	"Fences": {
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
	"Fortress": {
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
	"Four Towers": {
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
	"Ice": {
		Name:          "Ice",
		BoostlessTime: 17.0,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      14.5,
				BoostTime: 2.5,
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
	"Ladder Slide": {
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
	"Ladder Tower": {
		Name:          "Ladder Tower",
		BoostlessTime: 25.25,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      13.0,
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
	"Overhead 4b": {
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
	"Quartz Climb": {
		Name:          "Quartz Climb",
		BoostlessTime: 19.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      1.5,
				BoostTime: 14.5,
				Quality:   BestMove,
			},
			{
				Name:      "cp 1-2",
				Time:      11.5,
				BoostTime: 13.5,
				Quality:   BestMove,
			},
		},
	},
	"Quartz Temple": {
		Name:          "Quartz Temple",
		BoostlessTime: 16.25,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      8.5,
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
	"Rng Skip": {
		Name:          "Rng Skip",
		BoostlessTime: 11.75,
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
	"Sandpit": {
		// 13.5 15.5
		Name:          "Sandpit",
		BoostlessTime: 34.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      24.5,
				BoostTime: 3.5,
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
	"Scatter": {
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
	"Slime Scatter": {
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
	"Slime Skip": {
		Name:          "Slime Skip",
		BoostlessTime: 15.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      11.25,
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
	"Tower Tightrope": {
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
	"Triple Platform": {
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
	"Triple Trapdoors": {
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
	"Underbridge": {
		Name:          "Underbridge",
		BoostlessTime: 23.75,
		BoostStrats: []boostRoom{
			{
				Name:      "cp 0-1",
				Time:      20.5,
				BoostTime: 5.5,
				Quality:   BrilliantMove,
			},
			{
				Name:      "cp 1-2",
				Time:      10.5,
				BoostTime: 8.5,
				Quality:   BestMove,
			},
		},
	},
	"Finish Room": {
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
	res := make([]string, len(roomMap)-1)
	i := 0
	for _, v := range roomMap {
		if v.Name == "Finish Room" {
			continue
		}

		res[i] = v.Name
		i++
	}

	return res
}

func calcBoostless(roomList []string) float64 {
	time := 0.0
	for _, room := range roomList {
		time += roomMap[room].BoostlessTime
	}

	return time
}

type calcResultBoost struct {
	ind      int
	stratInd int
	pacelock float64
}

func calcTwoBoost(roomList []string) (float64, []calcResultBoost, error) {
	if roomMap[roomList[len(roomList)-1]].Name != "Finish Room" {
		err := fmt.Errorf("last room is supposed to be finish room. this is a programming error")
		log.Warn(err)
		return 0, nil, err
	}

	bestBoostTime := 600.0
	boostRooms := []calcResultBoost{
		{
			ind:      -1,
			stratInd: -1,
		},
		{
			ind:      -1,
			stratInd: -1,
		},
	}

	boostlessTime := calcBoostless(roomList)

	for i := 0; i < 9; i++ {
		for j := i + 1; j < 9; j++ {
			firstBoostRoom := roomMap[roomList[i]]
			secondBoostRoom := roomMap[roomList[j]]

			timeBetweenBoosts := 0.0
			for k := i + 1; k < j; k++ {
				timeBetweenBoosts += roomMap[roomList[k]].BoostlessTime
			}

			for firstBoostStrat := 0; firstBoostStrat < len(firstBoostRoom.BoostStrats); firstBoostStrat++ {
				for secondBoostStrat := 0; secondBoostStrat < len(secondBoostRoom.BoostStrats); secondBoostStrat++ {
					pacelock := max(0, 60-(timeBetweenBoosts+firstBoostRoom.BoostStrats[firstBoostStrat].Time-firstBoostRoom.BoostStrats[firstBoostStrat].BoostTime+secondBoostRoom.BoostStrats[secondBoostStrat].BoostTime))

					boostTime := boostlessTime - (firstBoostRoom.BoostlessTime - firstBoostRoom.BoostStrats[firstBoostStrat].Time) - (secondBoostRoom.BoostlessTime - secondBoostRoom.BoostStrats[secondBoostStrat].Time) + pacelock

					if boostTime < bestBoostTime {
						log.Tracef("boostless: %v", boostlessTime)
						log.Tracef("timesave: %v", (firstBoostRoom.BoostlessTime-firstBoostRoom.BoostStrats[firstBoostStrat].Time)+(secondBoostRoom.BoostlessTime-secondBoostRoom.BoostStrats[secondBoostStrat].Time)-pacelock)
						log.Tracef("pacelock: %v", pacelock)

						bestBoostTime = boostTime
						boostRooms = []calcResultBoost{
							{
								ind:      i,
								stratInd: firstBoostStrat,
							},
							{
								ind:      j,
								stratInd: secondBoostStrat,
								pacelock: pacelock,
							},
						}
					}
				}
			}
		}
	}

	return bestBoostTime, boostRooms, nil
}

func calcThreeBoost(roomList []string) (float64, []calcResultBoost, error) {
	if roomMap[roomList[len(roomList)-1]].Name != "Finish Room" {
		err := fmt.Errorf("last room is supposed to be finish room. this is a programming error")
		log.Warn(err)
		return 0, nil, err
	}

	bestBoostTime := 600.0
	boostRooms := []calcResultBoost{
		{
			ind:      -1,
			stratInd: -1,
		},
		{
			ind:      -1,
			stratInd: -1,
		},
		{
			ind:      -1,
			stratInd: -1,
		},
	}

	boostlessTime := calcBoostless(roomList)

	for i := 0; i < 9; i++ {
		for j := i + 1; j < 9; j++ {
			for k := j + 1; k < 9; k++ {
				firstBoostRoom := roomMap[roomList[i]]
				secondBoostRoom := roomMap[roomList[j]]
				thirdBoostRoom := roomMap[roomList[k]]

				timeBetweenBoosts12 := 0.0
				for m := i + 1; m < j; m++ {
					timeBetweenBoosts12 += roomMap[roomList[m]].BoostlessTime
				}
				timeBetweenBoosts23 := 0.0
				for m := j + 1; m < k; m++ {
					timeBetweenBoosts23 += roomMap[roomList[m]].BoostlessTime
				}

				for firstBoostStrat := 0; firstBoostStrat < len(firstBoostRoom.BoostStrats); firstBoostStrat++ {
					for secondBoostStrat := 0; secondBoostStrat < len(secondBoostRoom.BoostStrats); secondBoostStrat++ {
						for thirdBoostStrat := 0; thirdBoostStrat < len(thirdBoostRoom.BoostStrats); thirdBoostStrat++ {
							pacelock1 := max(0, 60-(timeBetweenBoosts12+firstBoostRoom.BoostStrats[firstBoostStrat].Time-firstBoostRoom.BoostStrats[firstBoostStrat].BoostTime+secondBoostRoom.BoostStrats[secondBoostStrat].BoostTime))

							pacelock2 := max(0, 60-(timeBetweenBoosts23+secondBoostRoom.BoostStrats[secondBoostStrat].Time-secondBoostRoom.BoostStrats[secondBoostStrat].BoostTime+thirdBoostRoom.BoostStrats[thirdBoostStrat].BoostTime))
							boostTime := boostlessTime - (firstBoostRoom.BoostlessTime - firstBoostRoom.BoostStrats[firstBoostStrat].Time) - (secondBoostRoom.BoostlessTime - secondBoostRoom.BoostStrats[secondBoostStrat].Time) - (thirdBoostRoom.BoostlessTime - thirdBoostRoom.BoostStrats[thirdBoostStrat].Time) + pacelock1 + pacelock2

							// if i == 0 && j == 4 && k == 7 {
							// 	log.Info("")
							// 	log.Infof("%v %v %v", firstBoostStrat, secondBoostStrat, thirdBoostStrat)
							// 	log.Infof("pacelock 1: %v", pacelock1)
							// 	log.Infof("pacelock 2: %v", pacelock2)
							// 	log.Infof("boostless: %v", boostlessTime)
							// 	log.Infof("boost: %v", boostTime)
							// }

							if boostTime < bestBoostTime {
								log.Tracef("boostless: %v", boostlessTime)
								// log.Tracef("timesave: %v", (firstBoostRoom.BoostlessTime-firstBoostRoom.BoostStrats[firstBoostStrat].Time)+(secondBoostRoom.BoostlessTime-secondBoostRoom.BoostStrats[secondBoostStrat].Time)-pacelock)
								log.Tracef("pacelock 1: %v", pacelock1)
								log.Tracef("pacelock 2: %v", pacelock2)

								bestBoostTime = boostTime
								boostRooms = []calcResultBoost{
									{
										ind:      i,
										stratInd: firstBoostStrat,
									},
									{
										ind:      j,
										stratInd: secondBoostStrat,
										pacelock: pacelock1,
									},
									{
										ind:      k,
										stratInd: thirdBoostStrat,
										pacelock: pacelock2,
									},
								}
							}
						}
					}
				}
			}
		}
	}

	return bestBoostTime, boostRooms, nil
}

type calcSeedResult struct {
	boostlessTime float64
	boostTime     float64
	boostRooms    []calcResultBoost
}

func calcSeedInternal(roomList []string) (calcSeedResult, error) {
	res := calcSeedResult{boostlessTime: calcBoostless(roomList)}

	twoBoost, twoBoostRooms, err := calcTwoBoost(roomList)
	if err != nil {
		log.Warn(err)
		return calcSeedResult{}, err
	}

	threeBoost, threeBoostRooms, err := calcThreeBoost(roomList)
	if err != nil {
		log.Warn(err)
		return calcSeedResult{}, err
	}

	if twoBoost < threeBoost {
		res.boostTime = twoBoost
		res.boostRooms = twoBoostRooms

		return res, nil
	}

	res.boostTime = threeBoost
	res.boostRooms = threeBoostRooms
	return res, nil
}

func CalcSeed(roomList []string) (bytes.Buffer, error) {
	width, height := 775, 490
	dc := gg.NewContext(width, height)

	bgFile, err := os.Open("images/background.png")
	if err != nil {
		log.Warn(err)
		return bytes.Buffer{}, err
	}
	defer bgFile.Close()

	bgImage, _, err := image.Decode(bgFile)
	if err != nil {
		log.Warn(err)
		return bytes.Buffer{}, err
	}

	bgWidth := bgImage.Bounds().Dx()
	bgHeight := bgImage.Bounds().Dy()
	scaleX := float64(width) / float64(bgWidth)
	scaleY := float64(height) / float64(bgHeight)
	scale := math.Max(scaleX, scaleY)

	newWidth := float64(bgWidth) * scale
	newHeight := float64(bgHeight) * scale
	x := (float64(width) - newWidth) / 2
	y := (float64(height) - newHeight) / 2

	dc.Push()
	dc.Scale(scale, scale)
	dc.DrawImage(bgImage, int(x/scale), int(y/scale))
	dc.Pop()

	if err := dc.LoadFontFace("font/minecraft_font.ttf", 24); err != nil {
		log.Warn(err)
		return bytes.Buffer{}, err
	}

	type RoomInfo struct {
		text        string
		highlight   bool
		checkpoint  string
		pacelock    string
		moveQuality MoveQuality
	}

	roomList = append(roomList, "Finish Room")
	res, err := calcSeedInternal(roomList)
	if err != nil {
		log.Warn(err)
		return bytes.Buffer{}, err
	}

	roomsOutput := make([]RoomInfo, 0, 9)
	for i := 0; i < 9; i++ {
		roomsOutput = append(roomsOutput, RoomInfo{
			text: roomList[i],
		})
	}

	for _, br := range res.boostRooms {
		roomsOutput[br.ind].highlight = true
		roomsOutput[br.ind].checkpoint = roomMap[roomList[br.ind]].BoostStrats[br.stratInd].Name
		roomsOutput[br.ind].moveQuality = roomMap[roomList[br.ind]].BoostStrats[br.stratInd].Quality
		if math.Abs(br.pacelock) >= 1e-6 {
			roomsOutput[br.ind].pacelock = fmt.Sprintf("pacelock %vs", br.pacelock)
		}

	}

	if !roomsOutput[len(roomsOutput)-1].highlight {
		roomsOutput = roomsOutput[:8]
	}

	// Calculate maximum text width for consistent rectangle size
	var maxWidth float64
	for _, stat := range roomsOutput {
		displayText := stat.text
		if stat.highlight {
			displayText = fmt.Sprintf("%s (%s)", stat.text, stat.checkpoint)
		}
		width, _ := dc.MeasureString(displayText)
		if width > maxWidth {
			maxWidth = width
		}
	}

	rectWidth := maxWidth + 40
	rectHeight := float64(30)

	y = 40

	for _, room := range roomsOutput {
		rectX := float64(width)/2 - rectWidth/2
		rectY := float64(y) - rectHeight/2

		if room.highlight {
			// First draw background with move quality color
			dc.Push()
			switch room.moveQuality {
			case BrilliantMove:
				dc.SetRGBA(float64(brilliantMoveColor.R)/255,
					float64(brilliantMoveColor.G)/255,
					float64(brilliantMoveColor.B)/255,
					float64(brilliantMoveColor.A)/255)
			case GreatMove:
				dc.SetRGBA(float64(greatMoveColor.R)/255,
					float64(greatMoveColor.G)/255,
					float64(greatMoveColor.B)/255,
					float64(greatMoveColor.A)/255)
			default: // BestMove
				dc.SetRGBA(float64(bestMoveColor.R)/255,
					float64(bestMoveColor.G)/255,
					float64(bestMoveColor.B)/255,
					float64(bestMoveColor.A)/255)
			}
			dc.DrawRoundedRectangle(rectX, rectY, rectWidth, rectHeight, 10)
			dc.Fill()
			dc.Pop()

			// Draw icon based on move quality
			switch room.moveQuality {
			case BrilliantMove:
				if img, err := gg.LoadImage("images/brilliant.png"); err == nil {
					iconSize := rectHeight
					imgHeight := float64(img.Bounds().Dy())
					scale := iconSize / imgHeight

					dc.Push()
					dc.Translate(float64(int(rectX-25)), float64(y))
					dc.Scale(scale, scale)
					dc.DrawImageAnchored(img, 0, 0, 0.5, 0.5)
					dc.Pop()
				}
			case GreatMove:
				if img, err := gg.LoadImage("images/great.png"); err == nil {
					iconSize := rectHeight
					imgHeight := float64(img.Bounds().Dy())
					scale := iconSize / imgHeight

					dc.Push()
					dc.Translate(float64(int(rectX-25)), float64(y))
					dc.Scale(scale, scale)
					dc.DrawImageAnchored(img, 0, 0, 0.5, 0.5)
					dc.Pop()
				}
			default: // BestMove
				if img, err := gg.LoadImage("images/best.png"); err == nil {
					iconSize := rectHeight
					imgHeight := float64(img.Bounds().Dy())
					scale := iconSize / imgHeight

					dc.Push()
					dc.Translate(float64(int(rectX-25)), float64(y))
					dc.Scale(scale, scale)
					dc.DrawImageAnchored(img, 0, 0, 0.5, 0.5)
					dc.Pop()
				}
			}

			// Draw text in white
			dc.SetColor(color.White)
			displayText := fmt.Sprintf("%s (%s)", room.text, room.checkpoint)
			dc.DrawStringAnchored(displayText, float64(width)/2, float64(y), 0.5, 0.5)

			dc.SetColor(color.RGBA{255, 255, 200, 255})
			dc.DrawStringAnchored(room.pacelock,
				rectX+rectWidth+20,
				float64(y),
				0, 0.5)
		} else {
			dc.Push()
			dc.SetRGBA(0, 0, 0, 0.5)
			dc.DrawRoundedRectangle(rectX, rectY, rectWidth, rectHeight, 10)
			dc.Fill()
			dc.Pop()

			dc.SetColor(color.White)
			dc.DrawStringAnchored(room.text, float64(width)/2, float64(y), 0.5, 0.5)
		}

		y += 40
	}
	y += 20

	timeTexts := []struct {
		prefix string
		time   string
	}{
		{"Boost time: ", formatTime(res.boostTime)},
		{"Boostless time: ", formatTime(res.boostlessTime)},
	}

	var maxPrefixWidth, maxTimeWidth float64
	for _, tt := range timeTexts {
		prefixWidth, _ := dc.MeasureString(tt.prefix)
		timeWidth, _ := dc.MeasureString(tt.time)
		if prefixWidth > maxPrefixWidth {
			maxPrefixWidth = prefixWidth
		}
		if timeWidth > maxTimeWidth {
			maxTimeWidth = timeWidth
		}
	}

	totalWidth := maxPrefixWidth + maxTimeWidth
	startX := float64(width)/2 - totalWidth/2

	dc.SetColor(color.White)
	for _, tt := range timeTexts {
		dc.DrawString(tt.prefix, startX, float64(y))
		dc.DrawString(tt.time, startX+maxPrefixWidth, float64(y))
		y += 30
	}

	var buf bytes.Buffer
	dc.EncodePNG(&buf)
	return buf, nil
}

func formatTime(seconds float64) string {
	decimal := seconds - float64(int(seconds))
	if decimal != 0.0 && decimal != 0.5 {
		if decimal < 0.5 {
			seconds = float64(int(seconds)) + 0.5
		} else {
			seconds = float64(int(seconds)) + 1.0
		}
	}

	minutes := int(seconds) / 60
	remainingSeconds := seconds - float64(minutes*60)

	if minutes > 0 {
		return fmt.Sprintf("%d:%04.1f", minutes, remainingSeconds)
	}
	return fmt.Sprintf("%.1f", remainingSeconds)
}
