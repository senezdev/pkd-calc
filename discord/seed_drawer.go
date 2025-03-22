package discord

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strings"

	"pkd-bot/calc"

	"github.com/fogleman/gg"
	log "github.com/sirupsen/logrus"
)

var (
	bestMoveColor      = color.RGBA{155, 199, 0, 200}
	greatMoveColor     = color.RGBA{0, 121, 211, 200}
	brilliantMoveColor = color.RGBA{48, 162, 197, 200}
)

func drawCalcResults(roomList []string, calcResults []calc.CalcSeedResult) (bytes.Buffer, error) {
	if roomList[len(roomList)-1] != "finish room" {
		roomList = append(roomList, "finish room")
	}

	tempDC := gg.NewContext(1, 1)
	if err := tempDC.LoadFontFace("font/minecraft_font.ttf", 24); err != nil {
		log.Warn(err)
		return bytes.Buffer{}, err
	}

	res := calcResults[0]
	maxPacelockWidth := 0.0
	for _, br := range res.BoostRooms {
		if math.Abs(br.Pacelock) >= 1e-6 {
			roundedPacelock := math.Round(br.Pacelock*10) / 10
			pacelockText := fmt.Sprintf("pacelock %.1fs", roundedPacelock)
			width, _ := tempDC.MeasureString(pacelockText)
			if width > maxPacelockWidth {
				maxPacelockWidth = width
			}
		}
	}

	width := 775
	if maxPacelockWidth > 0 {
		width = 775 + int(maxPacelockWidth) + 40 // Add padding
	}
	height := 490

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
		moveQuality calc.MoveQuality
	}

	roomsOutput := make([]RoomInfo, 0, 9)
	for i := 0; i < 9; i++ {
		words := strings.Split(roomList[i], " ")
		for j := range words {
			if len(words[j]) > 0 {
				words[j] = strings.ToUpper(string(words[j][0])) + words[j][1:]
			}
		}
		roomsOutput = append(roomsOutput, RoomInfo{
			text: strings.Join(words, " "),
		})
	}

	for _, br := range res.BoostRooms {
		roomsOutput[br.Ind].highlight = true
		roomsOutput[br.Ind].checkpoint = calc.RoomMap[roomList[br.Ind]].BoostStrats[br.StratInd].Name
		roomsOutput[br.Ind].moveQuality = calc.RoomMap[roomList[br.Ind]].BoostStrats[br.StratInd].Quality
		if math.Abs(br.Pacelock) >= 1e-6 {
			roomsOutput[br.Ind].pacelock = fmt.Sprintf("pacelock %.1fs", math.Round(br.Pacelock*10)/10)
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
			case calc.BrilliantMove:
				dc.SetRGBA(float64(brilliantMoveColor.R)/255,
					float64(brilliantMoveColor.G)/255,
					float64(brilliantMoveColor.B)/255,
					float64(brilliantMoveColor.A)/255)
			case calc.GreatMove:
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
			case calc.BrilliantMove:
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
			case calc.GreatMove:
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
			var displayText string
			displayText = fmt.Sprintf("%s (%s)", room.text, room.checkpoint)
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

			var displayText string
			displayText = room.text

			dc.DrawStringAnchored(displayText, float64(width)/2, float64(y), 0.5, 0.5)
		}

		y += 40
	}
	y += 20

	timeTexts := []struct {
		prefix string
		time   string
	}{
		{"Boost time: ", FormatTime(res.BoostTime)},
		{"Boostless time: ", FormatTime(res.BoostlessTime)},
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

func FormatTime(seconds float64) string {
	minutes := int(seconds) / 60
	remainingSeconds := seconds - float64(minutes*60)

	if minutes > 0 {
		return fmt.Sprintf("%d:%04.1f", minutes, remainingSeconds)
	}
	return fmt.Sprintf("%.1f", remainingSeconds)
}
