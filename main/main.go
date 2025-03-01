package main

import (
	"pkd-bot/discord"
	"pkd-bot/server"

	log "github.com/sirupsen/logrus"
)

func main() {
	go func() {
		if err := server.StartServer(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := discord.StartDiscordBot(); err != nil {
		log.Fatal(err)
	}
}
