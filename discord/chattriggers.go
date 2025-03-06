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
	switch ign {
	case "Tauktes":
		ign = "PooPooPooPoo"
	case "Blrk":
		ign = "0Zl4Ms9Jc3_B8aN6"
	case "O_N_E_Dimension":
		ign = "P_O_O_Dimension"
	case "senez":
		ign = "airh4ck"
	}

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

	// Generate a unique key for this seed based on the room combination
	seedKey := strings.Join(rooms, "|")

	// Check if we should send a Discord message (if boost time is good and we haven't seen this seed recently)
	if bestResult.BoostTime < 130 && !seedCache.HasSeen(seedKey) && !debug {
		// Mark this seed as seen to prevent duplicate messages
		seedCache.MarkSeen(seedKey)

		img, err := drawCalcResults(rooms, []calc.CalcSeedResult{bestResult})
		if err != nil {
			return calc.CalcSeedResult{}, nil, fmt.Errorf("error drawing seed results: %w", err)
		}

		content := fmt.Sprintf("%s has found a %s seed, %s requeues in %s",
			ign, FormatTime(bestResult.BoostTime), lobby, timeLeft)

		_, err = s.ChannelMessageSendComplex(BotCommandsChannelID, &discordgo.MessageSend{
			Content: content,
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
