package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	TOKEN      string
	APP_ID     string
	GUILD_ID   string
	CHANNEL_ID string
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Failed to load env variables")
	}
	TOKEN = os.Getenv("DISCORD_TOKEN")
	APP_ID = os.Getenv("APP_ID")
	GUILD_ID = os.Getenv("GUILD_ID")
	CHANNEL_ID = os.Getenv("CHANNEL_ID")

	discord, err := discordgo.New("Bot " + TOKEN)
	if err != nil {
		panic("Failed to start discord session")
	}

	// executes on new interaction
	discord.AddHandler(newInteraction)

	// executes on startup
	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %s", r.User.String())
	})

	// updates the list of active slash(/) commands on discord
	// command declarations are in commands.go
	_, err = discord.ApplicationCommandBulkOverwrite(APP_ID, GUILD_ID, commands)
	if err != nil {
		log.Fatalf("could not register commands: %s", err)
	}

	// We only really care about receiving voice state updates.
	discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildVoiceStates)
	err = discord.Open()

	if err != nil {
		log.Fatalf("could not open session: %s", err)
	}

	// wait for CTRL + C then shutdown
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	err = discord.Close()
	if err != nil {
		log.Printf("could not close session gracefully: %s", err)
	}
}
