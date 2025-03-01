package discord

import (
	"bytes"
	"fmt"

	"pkd-bot/calc"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

var BotCommandsChannelID = ""

func ChattriggersHandle(rooms []string, timeLeft, lobby string, debug bool) (calc.CalcSeedResult, error) {
	if s == nil {
		return calc.CalcSeedResult{}, fmt.Errorf("discord session is not initialized")
	}

	if BotCommandsChannelID == "" {
		BotCommandsChannelID = GetChannelIDByName("bot-commands")
		if BotCommandsChannelID == "" {
			return calc.CalcSeedResult{}, fmt.Errorf("could not find #bot-commands channel")
		}
	}

	if err := checkBotPermissions(BotCommandsChannelID); err != nil {
		return calc.CalcSeedResult{}, fmt.Errorf("permission error: %w", err)
	}

	results, err := calc.CalcSeed(rooms)
	if err != nil {
		return calc.CalcSeedResult{}, fmt.Errorf("error calculating seed: %w", err)
	}

	if len(results) == 0 {
		return calc.CalcSeedResult{}, fmt.Errorf("no results found for the given rooms")
	}

	bestResult := results[0]

	if bestResult.BoostTime < 130 && !debug {
		img, err := drawCalcResults(rooms, []calc.CalcSeedResult{bestResult})
		if err != nil {
			return calc.CalcSeedResult{}, fmt.Errorf("error drawing seed results: %w", err)
		}

		content := fmt.Sprintf("A player has found a %s seed, %s requeues in %s",
			formatTime(bestResult.BoostTime), lobby, timeLeft)

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
			return calc.CalcSeedResult{}, fmt.Errorf("error sending message to Discord: %w", err)
		}
	}

	return bestResult, nil
}

func GetChannelIDByName(channelName string) string {
	if s == nil {
		log.Error("Discord session is not initialized")
		return ""
	}

	// If GuildID is empty, we need to search through all available guilds
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
