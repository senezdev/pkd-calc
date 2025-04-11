package discord

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"pkd-bot/calc"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

var BotCommandsChannelID = ""

var ct2blrk = map[string]string{
	"Early 3-1":   "Early 3+1",
	"Glass Neo":   "Rng Skip",
	"Overhead 4B": "Overhead 4b",
}

type BoostRoomsResponse struct {
	Name     string  `json:"name"`
	Pacelock float64 `json:"pacelock"`
	Index    int     `json:"index"`
}

var seedCache = NewSeedCache(1 * time.Hour)

func ChattriggersHandle(rooms []string, timeLeft, lobby, ign string, debug bool) (calc.CalcSeedResult, []BoostRoomsResponse, error) {
	if s == nil {
		return calc.CalcSeedResult{}, nil, fmt.Errorf("discord session is not initialized")
	}

	for i, r := range rooms {
		blrkRoom, exists := ct2blrk[r]
		if exists {
			rooms[i] = blrkRoom
		}

		rooms[i] = strings.ToLower(rooms[i])
	}
	rooms = append(rooms, "finish room")

	if BotCommandsChannelID == "" {
		BotCommandsChannelID = GetChannelIDByName("bot-commands")
		if BotCommandsChannelID == "" {
			return calc.CalcSeedResult{}, nil, fmt.Errorf("could not find #bot-commands channel")
		}
	}

	if err := checkBotPermissions(BotCommandsChannelID); err != nil {
		return calc.CalcSeedResult{}, nil, fmt.Errorf("permission error: %w", err)
	}

	results, err := calc.CalcSeed(rooms)
	if err != nil {
		return calc.CalcSeedResult{}, nil, fmt.Errorf("error calculating seed: %w", err)
	}

	if len(results) == 0 {
		return calc.CalcSeedResult{}, nil, fmt.Errorf("no results found for the given rooms")
	}

	bestResult := results[0]

	seedKey := strings.Join(rooms, "|")

	if bestResult.BoostTime < 130 && !seedCache.HasSeen(seedKey) && !debug {
		seedCache.MarkSeen(seedKey)

		img, err := drawCalcResults(rooms, []calc.CalcSeedResult{bestResult})
		if err != nil {
			return calc.CalcSeedResult{}, nil, fmt.Errorf("error drawing seed results: %w", err)
		}

		content := fmt.Sprintf("%s has found a %s seed, %s requeues in %s",
			ign, FormatTime(bestResult.BoostTime), lobby, timeLeft)

		calcCommand := createCalcCommand(rooms[:len(rooms)-1]) // Exclude "finish room"

		components := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						CustomID: ButtonShowCalc,
						Label:    "How did you get this?",
						Style:    discordgo.SuccessButton,
					},
					discordgo.Button{
						CustomID: ButtonCopyCalcCommand,
						Label:    "Copy Calc Command",
						Style:    discordgo.PrimaryButton,
						Emoji: &discordgo.ComponentEmoji{
							Name: "📋",
						},
					},
				},
			},
		}

		message, err := s.ChannelMessageSendComplex(BotCommandsChannelID, &discordgo.MessageSend{
			Content:    content,
			Components: components,
			Files: []*discordgo.File{
				{
					Name:   "seed.png",
					Reader: bytes.NewReader(img.Bytes()),
				},
			},
		})
		if err != nil {
			return calc.CalcSeedResult{}, nil, fmt.Errorf("error sending message to Discord: %w", err)
		}

		messageStates[message.ID] = &ResultState{
			Rooms:       rooms[:len(rooms)-1],
			Results:     []calc.CalcSeedResult{bestResult},
			Index:       0,
			Filter:      ButtonAnyBoost,
			CalcCommand: calcCommand, // Store the calc command in the state
		}

		cleanupTimers[message.ID] = cleanupMessageState(message.ID, s, BotCommandsChannelID, true)
	}

	boostRooms := make([]BoostRoomsResponse, 0)
	for _, room := range bestResult.BoostRooms {
		boostRooms = append(boostRooms, BoostRoomsResponse{
			Name:     fmt.Sprintf("%s (%s)", calc.RoomMap[rooms[room.Ind]].Name, calc.RoomMap[rooms[room.Ind]].BoostStrats[room.StratInd].Name),
			Pacelock: room.Pacelock,
			Index:    room.Ind,
		})
	}

	return bestResult, boostRooms, nil
}

type PkdutilResult struct {
	Best struct {
		Result     calc.CalcSeedResult
		BoostRooms []BoostRoomsResponse
	}
	Personal struct {
		Result     calc.CalcSeedResult
		BoostRooms []BoostRoomsResponse
	}
}

func PkdutilsHandle(rooms []string, splits map[string]calc.Room) (PkdutilResult, error) {
	if s == nil {
		return PkdutilResult{}, fmt.Errorf("discord session is not initialized")
	}

	for i := range rooms {
		rooms[i] = strings.ToLower(rooms[i])
	}
	rooms = append(rooms, "finish room")

	// calc with calc splits first
	results, err := calc.CalcSeed(rooms)
	if err != nil {
		return PkdutilResult{}, fmt.Errorf("error calculating seed: %w", err)
	}

	if len(results) == 0 {
		return PkdutilResult{}, fmt.Errorf("no results found for the given rooms")
	}

	bestResult := results[0]

	boostRooms := make([]BoostRoomsResponse, 0)
	for _, room := range bestResult.BoostRooms {
		boostRooms = append(boostRooms, BoostRoomsResponse{
			Name:     fmt.Sprintf("%s (%s)", calc.RoomMap[rooms[room.Ind]].Name, calc.RoomMap[rooms[room.Ind]].BoostStrats[room.StratInd].Name),
			Pacelock: room.Pacelock,
			Index:    room.Ind,
		})
	}

	// calc with personal splits next
	personalResults, err := calc.CalcSeedCustom(rooms, splits)
	if err != nil {
		return PkdutilResult{}, fmt.Errorf("error calculating seed: %w", err)
	}
	log.Debugf("%+v", personalResults[0])

	if len(personalResults) == 0 {
		return PkdutilResult{}, fmt.Errorf("no results found for the given rooms")
	}

	personalResult := personalResults[0]

	personalBoostRooms := make([]BoostRoomsResponse, 0)
	for _, room := range personalResult.BoostRooms {
		personalBoostRooms = append(personalBoostRooms, BoostRoomsResponse{
			Name:     fmt.Sprintf("%s (%s)", calc.RoomMap[rooms[room.Ind]].Name, calc.RoomMap[rooms[room.Ind]].BoostStrats[room.StratInd].Name),
			Pacelock: room.Pacelock,
			Index:    room.Ind,
		})
	}

	return PkdutilResult{
		Best: struct {
			Result     calc.CalcSeedResult
			BoostRooms []BoostRoomsResponse
		}{bestResult, boostRooms},
		Personal: struct {
			Result     calc.CalcSeedResult
			BoostRooms []BoostRoomsResponse
		}{personalResult, personalBoostRooms},
	}, nil
}

func GetChannelIDByName(channelName string) string {
	if s == nil {
		log.Error("Discord session is not initialized")
		return ""
	}

	if GuildID == "" {
		guilds, err := s.UserGuilds(100, "", "", false)
		if err != nil {
			log.Errorf("Error getting user guilds: %v", err)
			return ""
		}

		for _, guild := range guilds {
			channels, err := s.GuildChannels(guild.ID)
			if err != nil {
				log.Errorf("Error getting channels for guild %s: %v", guild.ID, err)
				continue
			}

			for _, channel := range channels {
				if channel.Type == discordgo.ChannelTypeGuildText && channel.Name == channelName {
					return channel.ID
				}
			}
		}
	} else {
		channels, err := s.GuildChannels(GuildID)
		if err != nil {
			log.Errorf("Error getting channels for guild %s: %v", GuildID, err)
			return ""
		}

		for _, channel := range channels {
			if channel.Type == discordgo.ChannelTypeGuildText && channel.Name == channelName {
				return channel.ID
			}
		}
	}

	log.Errorf("Channel '%s' not found", channelName)
	return ""
}

func checkBotPermissions(channelID string) error {
	log.Info("checking bot permissions")

	if s == nil {
		err := fmt.Errorf("discord session is not initialized")
		log.Error(err)
		return err
	}

	_, err := s.Channel(channelID)
	if err != nil {
		err := fmt.Errorf("error getting channel info: %w", err)
		log.Error(err)
		return err
	}

	permissions, err := s.State.UserChannelPermissions(s.State.User.ID, channelID)
	if err != nil {
		err := fmt.Errorf("error getting permissions: %w", err)
		log.Error(err)
		return err
	}

	requiredPerms := discordgo.PermissionViewChannel |
		discordgo.PermissionSendMessages |
		discordgo.PermissionAttachFiles

	if permissions&int64(requiredPerms) != int64(requiredPerms) {
		return fmt.Errorf("bot lacks necessary permissions for channel %s. Has: %d, Needs: %d",
			channelID, permissions, requiredPerms)
	}

	return nil
}

func createCalcCommand(rooms []string) string {
	var commandParts []string
	commandParts = append(commandParts, "/calc")

	for i, room := range rooms {
		// Format room names for the command
		roomName := room
		// First letter uppercase for each room
		if len(roomName) > 0 {
			roomName = strings.ToUpper(roomName[:1]) + roomName[1:]
		}
		commandParts = append(commandParts, fmt.Sprintf("room_%d:%s", i+1, roomName))
	}

	return strings.Join(commandParts, " ")
}
