package discord

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"time"

	"pkd-bot/calc"
	"pkd-bot/hypixel"
	"pkd-bot/tournaments"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func StartDiscordBot() error {
	slices.Sort(options)

	log.SetReportCaller(true)
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Infof("Logged in as %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionApplicationCommandAutocomplete:
			autocompleteHandler(s, i)
		case discordgo.InteractionMessageComponent:
			buttonHandler(s, i)
		}
	})

	err := s.Open()
	if err != nil {
		log.Errorf("Cannot open the session: %v", err)
		return err
	}

	logBotPermissions()

	log.Info("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, GuildID, v)
		if err != nil {
			log.Errorf("Cannot create '%v' command: %v", v.Name, err)
			return err
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Info("Press Ctrl+C to exit")
	<-stop

	log.Info("Shutting down...")

	return nil
}

var (
	BotToken = ""
	GuildID  = ""
)

var s *discordgo.Session

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to open .env")
	}

	BotToken = os.Getenv("BOT_TOKEN")
	GuildID = os.Getenv("GUILD_ID")

	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}

	var err error
	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot token, couldn't initiate a session: %v", err)
	}
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "calc",
		Description: "Choose 8 rooms",
		Options:     generateOptions(),
	},
	{
		Name:        "playercount",
		Description: "Find out the current player count in PKD!",
	},
}

func tournamentHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		err := fmt.Errorf("expected interaction type to be InteractionApplicationCommand, but found %v", i.Type)
		log.Warn(err)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "I'm sorry, I broke down. Please tell my developer that he's an idiot and he'll fix me.",
			},
		})
		return
	}

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You sent an incomplete command.", // TODO: need to send suggestions to the user on what to do next
			},
		})
	}

	content := ""

	switch options[0].Name {
	case "register":
		registerTournamentHandler(s, i)
		return
	default:
		content = "There is no such command"
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func registerTournamentHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// First, respond asking for the CSV file
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Please provide a CSV file with the tournament participants in your next message.",
		},
	})
	if err != nil {
		log.Errorf("Failed to send initial response: %v", err)
		return
	}

	// Create a message handler to wait for the CSV file
	s.AddHandlerOnce(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		log.Debugf("%+v", m.Embeds)
		// Ensure it's from the same user and channel
		if m.Author.ID != i.Member.User.ID || m.ChannelID != i.ChannelID {
			return
		}

		// Check if there's an attachment
		if len(m.Attachments) == 0 {
			s.ChannelMessageSend(m.ChannelID, "Please attach a CSV file.")
			return
		}

		attachment := m.Attachments[0]
		if !strings.HasSuffix(strings.ToLower(attachment.Filename), ".csv") {
			s.ChannelMessageSend(m.ChannelID, "The attached file must be a CSV file.")
			return
		}

		resp, err := http.Get(attachment.URL)
		if err != nil {
			log.Warnf("Failed to download attachment: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Failed to download the attachment. Please try again.")
			return
		}
		defer resp.Body.Close()

		fileContent, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Failed to read attachment content: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Failed to read the attachment. Please try again.")
			return
		}

		if err := tournaments.RegisterTournamentFromCsv(fileContent); err != nil {
			log.Errorf("Failed to register tournament: %v", err)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to register tournament: %v", err))
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Tournament successfully registered!")
	})
}

func playercountHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	logUserInteraction(i, "command", "playercount")

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{},
	})
	if err != nil {
		log.Errorf("Failed to defer response: %v", err)
		return
	}

	playerCount, err := hypixel.GetPlayerCount()
	content := "Failed to fetch player count from Hypixel API. Please try again later."
	if err != nil {
		log.Errorf("Failed to get player count: %v", err)
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Errorf("Failed to edit response with error message: %v", err)
		}
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "Parkour Duels",
		Color: 0x45D3B3,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://static.wikia.nocookie.net/hypixel-skyblock/images/5/5c/Hypixel_Logo.png/revision/latest/scale-to-width-down/250",
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Current player count",
				Value:  fmt.Sprintf("**%d**", playerCount),
				Inline: false,
			},
		},
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		log.Errorf("Failed to edit response with player count: %v", err)
	}
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"basic-command": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Hey there! Congratulations, you just executed your first slash command",
			},
		})
	},
	"tournament":  tournamentHandler,
	"calc":        calcSeedHandler,
	"playercount": playercountHandler,
}

var options = calc.GetRooms()

func generateOptions() []*discordgo.ApplicationCommandOption {
	var params []*discordgo.ApplicationCommandOption
	for i := 1; i <= 8; i++ {
		params = append(params, &discordgo.ApplicationCommandOption{
			Type:         discordgo.ApplicationCommandOptionString,
			Name:         fmt.Sprintf("room_%d", i),
			Description:  fmt.Sprintf("Choose option for room %d", i),
			Required:     true,
			Autocomplete: true,
		})
	}
	return params
}

const (
	ButtonPrevious   = "previous"
	ButtonNext       = "next"
	ButtonTwoBoost   = "two_boost"
	ButtonThreeBoost = "three_boost"
	ButtonAnyBoost   = "any_boost"
	ButtonShowCalc   = "show_calculation"
)

func createNavigationButtons(currentIndex, totalResults int, currentFilter string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: ButtonPrevious,
					Style:    discordgo.SecondaryButton,
					Emoji: &discordgo.ComponentEmoji{
						Name: "⬅️",
					},
					Disabled: currentIndex <= 0,
				},
				discordgo.Button{
					CustomID: ButtonNext,
					Style:    discordgo.SecondaryButton,
					Emoji: &discordgo.ComponentEmoji{
						Name: "➡️",
					},
					Disabled: currentIndex >= totalResults-1,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: ButtonTwoBoost,
					Label:    "2 Boost",
					Style: func() discordgo.ButtonStyle {
						if currentFilter == ButtonTwoBoost {
							return discordgo.PrimaryButton
						}
						return discordgo.SecondaryButton
					}(),
				},
				discordgo.Button{
					CustomID: ButtonThreeBoost,
					Label:    "3 Boost",
					Style: func() discordgo.ButtonStyle {
						if currentFilter == ButtonThreeBoost {
							return discordgo.PrimaryButton
						}
						return discordgo.SecondaryButton
					}(),
				},
				discordgo.Button{
					CustomID: ButtonAnyBoost,
					Label:    "Any Boost",
					Style: func() discordgo.ButtonStyle {
						if currentFilter == ButtonAnyBoost {
							return discordgo.PrimaryButton
						}
						return discordgo.SecondaryButton
					}(),
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: ButtonShowCalc,
					Label:    "How did you get this?",
					Style:    discordgo.SuccessButton,
				},
			},
		},
	}
}

type ResultState struct {
	Rooms   []string
	Results []calc.CalcSeedResult
	Index   int
	Filter  string
}

var messageStates = make(map[string]*ResultState)

var cleanupTimers = make(map[string]*time.Timer)

var showCalcMessages = make(map[string]string)

var (
	showCalcTimers     = make(map[string]*time.Timer)
	longButtonDuration = 5 * time.Minute
)

// Modify cleanupMessageState to use a timer
func cleanupMessageState(messageID string, s *discordgo.Session, channelID string, keepShowCalcButton bool) *time.Timer {
	return time.AfterFunc(15*time.Second, func() {
		// Get the current message
		message, err := s.ChannelMessage(channelID, messageID)
		if err != nil {
			log.Errorf("Failed to get message for cleanup: %v", err)
			return
		}

		// Get the current image
		if len(message.Attachments) == 0 {
			log.Error("No attachments found in message")
			return
		}

		// Download the current image
		resp, err := http.Get(message.Attachments[0].URL)
		if err != nil {
			log.Errorf("Failed to get attachment: %v", err)
			return
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Failed to read attachment data: %v", err)
			return
		}

		var components []discordgo.MessageComponent
		if keepShowCalcButton {
			// Keep only the ShowCalc button
			components = []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							CustomID: ButtonShowCalc,
							Label:    "How did you get this?",
							Style:    discordgo.SuccessButton,
						},
					},
				},
			}

			// Create a timer to remove the ShowCalc button and state eventually
			showCalcTimers[messageID] = time.AfterFunc(5*time.Minute, func() {
				// Remove all buttons after the extended period
				_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
					ID:          messageID,
					Channel:     channelID,
					Files:       []*discordgo.File{{Name: "result.png", Reader: bytes.NewReader(data)}},
					Components:  &[]discordgo.MessageComponent{},
					Attachments: &[]*discordgo.MessageAttachment{},
				})
				if err != nil {
					log.Errorf("Failed to remove ShowCalc button: %v", err)
				}
				delete(showCalcTimers, messageID)
				delete(messageStates, messageID) // Only delete state when fully done
			})
		} else {
			components = []discordgo.MessageComponent{}
			// If not keeping the button, remove the state now
			delete(messageStates, messageID)
		}

		// Keep the last image but remove/modify buttons
		_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			ID:          messageID,
			Channel:     channelID,
			Files:       []*discordgo.File{{Name: "result.png", Reader: bytes.NewReader(data)}},
			Components:  &components,
			Attachments: &[]*discordgo.MessageAttachment{},
		})
		if err != nil {
			log.Errorf("Failed to update buttons: %v", err)
		}

		delete(messageStates, messageID)
		delete(showCalcMessages, messageID)
		delete(cleanupTimers, messageID)
		// NOTE: We don't delete messageStates here if keepShowCalcButton is true
	})
}

func buttonHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	logUserInteraction(i, "button click", i.MessageComponentData().CustomID)

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("application panicked while handling a request: %v", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "HA. Buttons not working.",
				},
			})
		}
	}()

	state, exists := messageStates[i.Message.ID]
	if !exists {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This interaction has expired. Please run the command again.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Acknowledge the interaction first
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Errorf("Failed to acknowledge interaction: %v", err)
		return
	}

	// Reset the cleanup timer
	if timer, exists := cleanupTimers[i.Message.ID]; exists {
		timer.Reset(15 * time.Second)
	}

	if timer, exists := showCalcTimers[i.Message.ID]; exists {
		timer.Reset(longButtonDuration)
	}

	if i.MessageComponentData().CustomID == ButtonShowCalc {
		var result calc.CalcSeedResult
		filteredResults := getFilteredResults(state)
		if len(filteredResults) > 0 {
			if state.Index < len(filteredResults) {
				result = filteredResults[state.Index]
			} else {
				result = filteredResults[0] // Fallback to first result if index is out of bounds
			}
		} else {
			result = state.Results[0] // Fallback to first result in original results
		}

		// Create detailed calculation message
		detailedCalc := formatDetailedCalculation(state.Rooms, result)

		// Check if we already have a calculation message for this interaction
		if calcMsgID, exists := showCalcMessages[i.Message.ID]; exists {
			// Edit the existing message instead of sending a new one
			_, err = s.ChannelMessageEdit(i.ChannelID, calcMsgID, detailedCalc)
			if err != nil {
				log.Errorf("Failed to edit calculation message: %v", err)
				// If edit fails (message might be deleted), remove from map and send a new one
				delete(showCalcMessages, i.Message.ID)
				msg, err := s.ChannelMessageSend(i.ChannelID, detailedCalc)
				if err == nil {
					showCalcMessages[i.Message.ID] = msg.ID
				} else {
					log.Errorf("Failed to send calculation details: %v", err)
				}
			}
		} else {
			// Send a new message with the calculation details
			msg, err := s.ChannelMessageSend(i.ChannelID, detailedCalc)
			if err != nil {
				log.Errorf("Failed to send calculation details: %v", err)
			} else {
				// Store the message ID for future references
				showCalcMessages[i.Message.ID] = msg.ID
			}
		}

		// Edit the original interaction response to confirm
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{})
		if err != nil {
			log.Errorf("Failed to edit interaction response: %v", err)
		}

		return
	}

	switch i.MessageComponentData().CustomID {
	case ButtonPrevious:
		if state.Index > 0 {
			state.Index--
		}
	case ButtonNext:
		if state.Index < len(getFilteredResults(state))-1 {
			state.Index++
		}

		// Check if the current result has a boost time >= boostless time
		filteredResults := getFilteredResults(state)
		if len(filteredResults) > 0 && state.Index < len(filteredResults) {
			result := filteredResults[state.Index]

			// Calculate boostless time
			boostlessTime := 0.0
			for _, room := range state.Rooms {
				roomInfo := calc.RoomMap[room]
				boostlessTime += roomInfo.BoostlessTime
			}

			// Get boost time from the result
			boostTime := result.BoostTime

			// If boost time >= boostless time, remove buttons and image and send text message
			if boostTime >= boostlessTime {
				response := "Just play boostless at this point bro like why do you even bother with boosts 🤔"

				// Update the message to remove components and image
				_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
					ID:          i.Message.ID,
					Channel:     i.ChannelID,
					Content:     &response,
					Components:  &[]discordgo.MessageComponent{},
					Attachments: &[]*discordgo.MessageAttachment{},
				})
				if err != nil {
					log.Errorf("Failed to update message: %v", err)
					return
				}

				// Delete state since we're done with this interaction
				delete(messageStates, i.Message.ID)
				delete(showCalcMessages, i.Message.ID) // Clean up calculation message reference
				if timer, exists := cleanupTimers[i.Message.ID]; exists {
					timer.Stop()
					delete(cleanupTimers, i.Message.ID)
				}
				if timer, exists := showCalcTimers[i.Message.ID]; exists {
					timer.Stop()
					delete(showCalcTimers, i.Message.ID)
				}

				// Confirm interaction is complete
				_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{})
				if err != nil {
					log.Errorf("Failed to edit interaction response: %v", err)
				}

				return
			}
		}
	case ButtonTwoBoost, ButtonThreeBoost, ButtonAnyBoost:
		state.Filter = i.MessageComponentData().CustomID
		state.Index = 0
	}

	// Get filtered results
	filteredResults := getFilteredResults(state)

	// Make sure we have results to display
	if len(filteredResults) == 0 {
		log.Error("No results available after filtering")
		return
	}

	// Draw new image for the current index
	currentResult := []calc.CalcSeedResult{filteredResults[state.Index]}
	img, err := drawCalcResults(state.Rooms, currentResult)
	if err != nil {
		log.Error(err)
		return
	}

	// Create navigation buttons with updated state
	navButtons := createNavigationButtons(state.Index, len(filteredResults), state.Filter)

	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:          i.Message.ID,
		Channel:     i.ChannelID,
		Files:       []*discordgo.File{{Name: "result.png", Reader: bytes.NewReader(img.Bytes())}},
		Components:  &navButtons,
		Attachments: &[]*discordgo.MessageAttachment{},
	})
	if err != nil {
		log.Errorf("Failed to update message: %v", err)
		return
	}

	// Follow up with the interaction to confirm it's complete
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{})
	if err != nil {
		log.Errorf("Failed to edit interaction response: %v", err)
	}

	// Update the state in our map
	messageStates[i.Message.ID] = state
}

func formatDetailedCalculation(rooms []string, result calc.CalcSeedResult) string {
	rooms = append(rooms, "finish room")

	var boostCalc, boostlessCalc strings.Builder

	maxRoomNameLength := 0
	for _, room := range rooms {
		if len(room) > maxRoomNameLength {
			maxRoomNameLength = len(room)
		}
	}

	boostCalc.WriteString("**Boost time calculation:**\n```\n")
	boostlessCalc.WriteString("**Boostless time calculation:**\n```\n")

	formatStr := "%-" + fmt.Sprintf("%d", maxRoomNameLength+1) + "s: %6.2f"

	boostTimeSum := 0.0
	boostlessTimeSum := 0.0

	boostRooms := make(map[int]calc.CalcResultBoost)
	for _, br := range result.BoostRooms {
		boostRooms[br.Ind] = br
	}

	for i, room := range rooms {
		roomInfo := calc.RoomMap[room]

		boostlessTime := roomInfo.BoostlessTime
		boostlessTimeSum += boostlessTime
		boostlessCalc.WriteString(fmt.Sprintf(formatStr, room, boostlessTime))

		var boostLine strings.Builder

		if boost, isBoost := boostRooms[i]; isBoost {
			boostTime := roomInfo.BoostStrats[boost.StratInd].Time
			boostStart := roomInfo.BoostStrats[boost.StratInd].BoostTime
			boostTimeSum += boostTime

			boostLine.WriteString(fmt.Sprintf(formatStr, room, boostStart))

			boostLine.WriteString(fmt.Sprintf(" (before boost) + %5.2f", boostTime-boostStart))

			if boost.Pacelock > 0 {
				boostLine.WriteString(fmt.Sprintf(" + %5.2f (pacelock)", boost.Pacelock))
				boostTimeSum += boost.Pacelock
			}
		} else {
			boostTime := roomInfo.BoostlessTime
			boostTimeSum += boostTime
			boostLine.WriteString(fmt.Sprintf(formatStr, room, boostTime))
		}

		boostCalc.WriteString(boostLine.String())

		if i == 0 {
			boostlessCalc.WriteString(fmt.Sprintf(" - %5.2f (accounting for r1)", 0.3))
			boostCalc.WriteString(fmt.Sprintf(" - %5.2f (accounting for r1)", 0.3))

			boostlessTimeSum -= 0.3
			boostTimeSum -= 0.3
		} else {
			prevRoom := rooms[i-1]

			if prevRoom == "four towers" {
				boostlessCalc.WriteString(fmt.Sprintf(" - %5.2f (four towers timesave)", 0.2))
				boostCalc.WriteString(fmt.Sprintf(" - %5.2f (four towers timesave)", 0.2))

				boostlessTimeSum -= 0.2
				boostTimeSum -= 0.2
			}

			if prevRoom == "sandpit" || prevRoom == "castle wall" {
				boostlessCalc.WriteString(fmt.Sprintf(" - %5.2f (ib hh)", 0.1))
				boostCalc.WriteString(fmt.Sprintf(" - %5.2f (ib hh)", 0.1))

				boostlessTimeSum -= 0.1
				boostTimeSum -= 0.1
			}

			if prevRoom == "early 3+1" {
				for _, boost := range boostRooms {
					if boost.Ind == i-1 && boost.StratInd == 1 {
						boostCalc.WriteString(fmt.Sprintf(" - %5.2f (early 3+1 boost timesave)", 0.5))

						boostTimeSum -= 0.5
						break
					}
				}
			}

			if prevRoom == "underbridge" {
				for _, boost := range boostRooms {
					if boost.Ind == i-1 && boost.StratInd == 1 {
						boostCalc.WriteString(fmt.Sprintf(" - %5.2f (underbridge boost timesave)", 0.2))

						boostTimeSum -= 0.2
						break
					}
				}
			}
		}

		if i < len(rooms)-1 {
			boostCalc.WriteString("\n")
			boostlessCalc.WriteString("\n")
		}
	}

	separatorLine := strings.Repeat("-", maxRoomNameLength+20)

	boostCalc.WriteString(fmt.Sprintf("\n%s\nTotal: %6.2f seconds = %s\n", separatorLine, boostTimeSum, FormatTime(boostTimeSum)))
	boostCalc.WriteString("```\n")

	boostlessCalc.WriteString(fmt.Sprintf("\n%s\nTotal: %6.2f seconds = %s\n", separatorLine, boostlessTimeSum, FormatTime(boostlessTimeSum)))
	boostlessCalc.WriteString("```")

	timeSaved := boostlessTimeSum - boostTimeSum
	var comparisonText string
	if timeSaved > 0 {
		comparisonText = fmt.Sprintf("**Time saved with boosts: %.2f seconds**", timeSaved)
	} else {
		comparisonText = fmt.Sprintf("**Warning: Boosts are slower by %.2f seconds than boostless!**", -timeSaved)
	}

	return boostCalc.String() + "\n" + boostlessCalc.String() + "\n\n" + comparisonText
}

func getFilteredResults(state *ResultState) []calc.CalcSeedResult {
	if state.Filter == ButtonAnyBoost {
		return state.Results
	}

	filteredResults := make([]calc.CalcSeedResult, 0)
	for _, result := range state.Results {
		boostCount := len(result.BoostRooms)
		switch state.Filter {
		case ButtonTwoBoost:
			if boostCount == 2 {
				filteredResults = append(filteredResults, result)
			}
		case ButtonThreeBoost:
			if boostCount == 3 {
				filteredResults = append(filteredResults, result)
			}
		}
	}
	return filteredResults
}

func validateInput(input []string) (bool, error) {
	log.Info(options)

	if len(input) != 8 {
		err := fmt.Errorf("Was expecting 8 rooms, got %d", len(input))
		log.Error(err)
		return false, err
	}

	correctedInput := make([]string, len(input))
	copy(correctedInput, input)

	for i, roomName := range input {
		if slices.Contains(options, roomName) {
			continue
		}

		bestMatch, score := fuzzyMatch(roomName, options)

		if score >= 0.6 {
			log.Infof("Autocorrected '%s' to '%s' (score: %.2f)", roomName, bestMatch, score)
			correctedInput[i] = bestMatch
		} else {
			err := fmt.Errorf("I don't know a room called \"%s\". Did you mean \"%s\"?", roomName, bestMatch)
			log.Error(err)
			return false, err
		}
	}

	seen := make(map[string]bool)
	for _, room := range correctedInput {
		if seen[room] {
			err := fmt.Errorf("Room '%s' appears more than once. Each room must be unique", room)
			log.Error(err)
			return false, err
		}
		seen[room] = true
	}

	copy(input, correctedInput)
	return true, nil
}

func fuzzyMatch(input string, options []string) (string, float64) {
	input = strings.ToLower(input)
	bestMatch := ""
	bestScore := 0.0

	for _, option := range options {
		// Calculate match score
		score := calculateSimilarity(input, option)

		// Also check if the input is a prefix or substring
		optionLower := strings.ToLower(option)
		if strings.HasPrefix(optionLower, input) {
			// Prefix matches get a bonus
			score += 0.2
		} else if strings.Contains(optionLower, input) {
			// Substring matches get a smaller bonus
			score += 0.1
		}

		// Words appearing in the same order bonus
		inputWords := strings.Fields(input)
		if len(inputWords) > 1 {
			allWordsFound := true
			lastIndex := -1

			for _, word := range inputWords {
				idx := strings.Index(optionLower, word)
				if idx == -1 || idx <= lastIndex {
					allWordsFound = false
					break
				}
				lastIndex = idx
			}

			if allWordsFound {
				score += 0.15
			}
		}

		// Cap at 1.0
		if score > 1.0 {
			score = 1.0
		}

		if score > bestScore {
			bestScore = score
			bestMatch = option
		}
	}

	return bestMatch, bestScore
}

// calculateSimilarity computes a similarity score between two strings
// using a combination of Levenshtein distance and other heuristics
func calculateSimilarity(a, b string) float64 {
	a = strings.ToLower(a)
	b = strings.ToLower(b)

	// If strings are identical, return perfect score
	if a == b {
		return 1.0
	}

	// Handle acronyms - if input might be an acronym of the target
	// For example "tp" might match "triple platform"
	if isAcronymOf(a, b) {
		return 0.8
	}

	// Calculate Levenshtein distance
	distance := levenshteinDistance(a, b)
	maxLen := float64(max(len(a), len(b)))

	// Convert distance to similarity score (0 to 1)
	return 1.0 - float64(distance)/maxLen
}

// isAcronymOf checks if a might be an acronym of b
func isAcronymOf(potentialAcronym, fullText string) bool {
	if len(potentialAcronym) <= 1 {
		return false
	}

	words := strings.Fields(fullText)
	if len(potentialAcronym) != len(words) {
		return false
	}

	for i, char := range potentialAcronym {
		if i >= len(words) {
			return false
		}

		if len(words[i]) == 0 || !strings.HasPrefix(strings.ToLower(words[i]), string(char)) {
			return false
		}
	}

	return true
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create a matrix
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill the matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func calcSeedHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	logUserInteraction(i, "command", "calc")

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("application panicked while handling a request: %v", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Either you misspelled some room or the developer's an idiot. If it's the latter, go contact him and he'll fix me.",
				},
			})
		}
	}()

	data := i.ApplicationCommandData()
	selected := make([]string, 0, 8)

	for _, option := range data.Options {
		selected = append(selected, option.StringValue())
	}

	valid, err := validateInput(selected)
	if !valid {
		log.Error(err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})

		return
	}

	res, err := calc.CalcSeed(selected)
	if err != nil {
		log.Error(err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Go tell the developer he's an idiot 'cause something's broken idk",
			},
		})
		return
	}

	initialResult := []calc.CalcSeedResult{res[0]}
	img, err := drawCalcResults(selected, initialResult)
	if err != nil {
		log.Error(err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Go tell the developer he's an idiot 'cause something's broken idk",
			},
		})
		return
	}

	// Send initial response
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Files: []*discordgo.File{
				{
					Name:   "result.png",
					Reader: bytes.NewReader(img.Bytes()),
				},
			},
			Components: createNavigationButtons(0, len(res), ButtonAnyBoost),
		},
	})
	if err != nil {
		log.Error(err)
		return
	}

	// Get the message ID from the response
	message, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		log.Errorf("Failed to get interaction response: %v", err)
		return
	}

	// Store state with message ID
	messageStates[message.ID] = &ResultState{
		Rooms:   selected,
		Results: res,
		Index:   0,
		Filter:  ButtonAnyBoost,
	}

	timer := cleanupMessageState(message.ID, s, message.ChannelID, true)
	cleanupTimers[message.ID] = timer
}

func autocompleteHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Debug("Autocomplete handler triggered")

	data := i.ApplicationCommandData()
	log.Debugf("Command data: %+v", data)

	selectedOptions := make(map[string]bool)
	for _, opt := range data.Options {
		if !opt.Focused {
			selectedOptions[opt.StringValue()] = true
		}
	}

	var focusedOption *discordgo.ApplicationCommandInteractionDataOption
	for _, opt := range data.Options {
		if opt.Focused {
			focusedOption = opt
			break
		}
	}

	if focusedOption == nil {
		log.Error("No focused option found")
		return
	}

	log.Debugf("Focused option: %+v", focusedOption)
	searchTerm := strings.ToLower(focusedOption.StringValue())

	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, opt := range options {
		if selectedOptions[opt] {
			continue
		}

		if strings.Contains(strings.ToLower(opt), searchTerm) || len(searchTerm) == 0 {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  opt,
				Value: opt,
			})
		}
	}

	log.Debugf("Sending %d choices", len(choices))
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		log.Errorf("Failed to respond with choices: %v", err)
	}
}

func logUserInteraction(i *discordgo.InteractionCreate, interactionType string, actionName string) {
	var username, nickname, userID string

	// Get user ID
	if i.User != nil {
		// DM context
		userID = i.User.ID
		username = i.User.Username
		nickname = username // No nicknames in DMs
	} else if i.Member != nil {
		// Guild context
		userID = i.Member.User.ID
		username = i.Member.User.Username

		// Get nickname if available
		if i.Member.Nick != "" {
			nickname = i.Member.Nick
		} else {
			nickname = username
		}
	} else {
		log.Warn("Could not identify user from interaction")
		return
	}

	// Log the interaction with user details
	log.Infof("User %s (%s, ID: %s) performed %s: %s",
		nickname, username, userID, interactionType, actionName)

	// If it's the ShowCalc button, log more details
	if interactionType == "button click" && actionName == ButtonShowCalc {
		log.Infof("User %s (%s) requested calculation details", nickname, userID)
	}
}

func logBotPermissions() {
	if s == nil || s.State == nil || s.State.User == nil {
		log.Error("Discord session or user is not initialized, cannot check permissions")
		return
	}

	log.Info("=== Checking Bot Permissions ===")
	botID := s.State.User.ID
	botUsername := s.State.User.Username

	// Map to translate permission bits to readable names
	permissionNames := map[int64]string{
		discordgo.PermissionViewChannel:        "View Channels",
		discordgo.PermissionSendMessages:       "Send Messages",
		discordgo.PermissionAttachFiles:        "Attach Files",
		discordgo.PermissionEmbedLinks:         "Embed Links",
		discordgo.PermissionReadMessageHistory: "Read Message History",
		discordgo.PermissionManageMessages:     "Manage Messages",
		discordgo.PermissionMentionEveryone:    "Mention Everyone",
		discordgo.PermissionManageChannels:     "Manage Channels",
		discordgo.PermissionManageRoles:        "Manage Roles",
		discordgo.PermissionKickMembers:        "Kick Members",
		discordgo.PermissionBanMembers:         "Ban Members",
		discordgo.PermissionAdministrator:      "Administrator",
	}

	// Check if GUILD_ID is set
	if GuildID == "" {
		log.Error("GUILD_ID is not set in environment variables, can't check permissions")
		return
	}

	// Get the specific guild
	guild, err := s.Guild(GuildID)
	if err != nil {
		log.Errorf("Could not get details for guild ID %s: %v", GuildID, err)
		return
	}

	log.Infof("Bot %s#%s (ID: %s) checking permissions in server: %s",
		botUsername, s.State.User.Discriminator, botID, guild.Name)

	// Get bot's roles in this guild
	botMember, err := s.GuildMember(GuildID, botID)
	if err != nil {
		log.Errorf("Could not get bot's member info in guild %s: %v", guild.Name, err)
		return
	}

	// Get all roles to find bot's roles
	roles, err := s.GuildRoles(GuildID)
	if err != nil {
		log.Errorf("Could not get roles for guild %s: %v", guild.Name, err)
		return
	}

	// Log bot's role details
	log.Info("Bot role details:")
	for _, role := range roles {
		for _, botRoleID := range botMember.Roles {
			if role.ID == botRoleID {
				log.Infof("  - Role: %s (ID: %s, Position: %d, Permissions: %d)",
					role.Name, role.ID, role.Position, role.Permissions)

				// Log human-readable permissions
				var permissionsList []string
				for bit, name := range permissionNames {
					if role.Permissions&int64(bit) != 0 {
						permissionsList = append(permissionsList, name)
					}
				}
				log.Infof("    Permissions: %s", strings.Join(permissionsList, ", "))
			}
		}
	}

	// Check permissions in specific channels
	channels, err := s.GuildChannels(GuildID)
	if err != nil {
		log.Errorf("Could not get channels for guild %s: %v", guild.Name, err)
		return
	}

	// Filter for text channels only
	var textChannels []*discordgo.Channel
	for _, channel := range channels {
		if channel.Type == discordgo.ChannelTypeGuildText {
			textChannels = append(textChannels, channel)
		}
	}

	log.Infof("Checking permissions in %d text channels", len(textChannels))

	// Find #bot-commands channel specifically
	var botCommandsChannel *discordgo.Channel
	for _, channel := range textChannels {
		if channel.Name == "bot-commands" {
			botCommandsChannel = channel
			break
		}
	}

	// First check the bot-commands channel if found
	if botCommandsChannel != nil {
		perms, err := s.State.UserChannelPermissions(botID, botCommandsChannel.ID)
		if err != nil {
			log.Errorf("Error getting permissions for #bot-commands: %v", err)
		} else {
			log.Infof("=== #bot-commands Channel (ID: %s) ===", botCommandsChannel.ID)
			logChannelPermissions(perms, permissionNames)

			// Also store this ID for later use
			BotCommandsChannelID = botCommandsChannel.ID
		}
	} else {
		log.Warning("No #bot-commands channel found in this guild!")
	}

	// Log permissions for all text channels
	for _, channel := range textChannels {
		// Skip if this is the bot-commands channel we already checked
		if botCommandsChannel != nil && channel.ID == botCommandsChannel.ID {
			continue
		}

		perms, err := s.State.UserChannelPermissions(botID, channel.ID)
		if err != nil {
			log.Errorf("Error getting permissions for channel %s: %v", channel.Name, err)
			continue
		}

		log.Infof("=== Channel: %s (ID: %s) ===", channel.Name, channel.ID)
		logChannelPermissions(perms, permissionNames)
	}

	log.Info("=== Permission Check Complete ===")
}

// Helper function to log channel permissions
func logChannelPermissions(perms int64, permissionNames map[int64]string) {
	// Check critical permissions individually
	criticalPerms := []int64{
		discordgo.PermissionViewChannel,
		discordgo.PermissionSendMessages,
		discordgo.PermissionAttachFiles,
	}

	for _, perm := range criticalPerms {
		if perms&perm != 0 {
			log.Infof("  ✅ Has permission: %s", permissionNames[perm])
		} else {
			log.Errorf("  ❌ MISSING CRITICAL PERMISSION: %s", permissionNames[perm])
		}
	}

	// Log all other permissions
	log.Info("  Other permissions:")
	for bit, name := range permissionNames {
		// Skip the ones we already checked
		if contains(criticalPerms, bit) {
			continue
		}

		if perms&bit != 0 {
			log.Infof("    ✓ %s", name)
		}
	}
}

// Helper function to check if a slice contains a value
func contains(slice []int64, val int64) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
