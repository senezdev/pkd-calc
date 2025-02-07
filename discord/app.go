package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"pkd-bot/calc"
	"pkd-bot/tournaments"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

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
		Name:        "basic-command",
		Description: "Basic command",
	},
	{
		Name:        "tournament",
		Description: "Commands related to tournament management",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "register",
				Description: "Register a tournament",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	},
	{
		Name:        "calc",
		Description: "Choose 8 rooms",
		Options:     generateOptions(),
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

		// Remove this handler after processing
		// s.RemoveHandler(s.HandlerRemove)
	})
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
	"tournament": tournamentHandler,
	"calc":       calcSeedHandler,
}

var options = calc.GetRooms()

func generateOptions() []*discordgo.ApplicationCommandOption {
	var params []*discordgo.ApplicationCommandOption
	for i := 1; i <= 8; i++ {
		params = append(params, &discordgo.ApplicationCommandOption{
			Type:         discordgo.ApplicationCommandOptionString,
			Name:         fmt.Sprintf("room_%d", i), // Use fmt.Sprintf instead of string conversion
			Description:  fmt.Sprintf("Choose option for room %d", i),
			Required:     true,
			Autocomplete: true,
		})
	}
	return params
}

func calcSeedHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	selected := make([]string, 0, 8)

	for _, option := range data.Options {
		selected = append(selected, option.StringValue())
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
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Files: []*discordgo.File{
				{
					Name:   "result.png",
					Reader: bytes.NewReader(res.Bytes()),
				},
			},
		},
	})
}

func autocompleteHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Debug("Autocomplete handler triggered")

	data := i.ApplicationCommandData()
	log.Debugf("Command data: %+v", data)

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
		if strings.Contains(strings.ToLower(opt), searchTerm) {
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

func main() {
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
		}
	})

	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Info("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, GuildID, v)
		if err != nil {
			log.Fatalf("Cannot create '%v' command: %v", v.Name, err)
		}

		registeredCommands[i] = cmd
	}

	defer s.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Info("Press Ctrl+C to exit")
	<-stop

	log.Info("Shutting down...")
}
