package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/royarchl/D2-server-bot/bot/database"
	"github.com/royarchl/D2-server-bot/bot/interactions"

	"github.com/bwmarrin/discordgo"
)

var BotToken string
var BotChannelId string

func Run() {
	session, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatal("Failed to create Discord session: ", err)
	}

	session.Identify.Intents = discordgo.IntentsGuildMembers

	if err = session.Open(); err != nil {
		log.Fatal("Failed to connect to Discord servers: ", err)
	}
	defer session.Close()

	fmt.Println("Bot running... (Press Ctrl-C to exit)")

	BotChannelId = os.Getenv("BOT_CHANNEL_ID")
	interactions.RolesChannelId = os.Getenv("ROLES_CHANNEL_ID")

	// Register commands and buttons (+handlers)
	interactions.GuildId = os.Getenv("GUILD_ID")
	interactions.RegisterInteractions(session)
	defer interactions.DeregisterCommands(session)

	// Connect database
	err = database.EstablishDatabaseConnection()
	if err != nil {
		log.Fatal("Failed to establish database connection: ", err)
	}
	defer database.Db.Close()

	// Register users
	guildMembers, err := session.GuildMembers(interactions.GuildId, "", 1000)
	database.RegisterServerMembers(guildMembers)

	// NOTE: Temporary test for role buttons
	// NOTE: Move this into a function that accepts a file path and returns a string.
	data, err := os.ReadFile("txt/timezones-message.txt")
	if err != nil {
		log.Println("Error reading file: ", err)
		return
	}
	timezoneMessage := string(data)

	data, err = os.ReadFile("txt/dlc-message.txt")
	if err != nil {
		log.Println("Error reading file: ", err)
		return
	}
	dlcMessage := string(data)

	data, err = os.ReadFile("txt/experience-message.txt")
	if err != nil {
		log.Println("Error reading file: ", err)
		return
	}
	experienceMessage := string(data)

	interactions.GenerateRolesMessage(session, interactions.TimezoneRoles, timezoneMessage)
	interactions.GenerateRolesMessage(session, interactions.RaidRoles, dlcMessage)
	interactions.GenerateRolesMessage(session, interactions.ExperienceRoles, experienceMessage)

	// Send the "bot online" ack message
	message_botOnline := "*I'm back! Oh, my head... It feels like I was squeezed into a tiny little box.*"
	_, err = session.ChannelMessageSend(BotChannelId, message_botOnline)
	if err != nil {
		log.Println("Failed to send startup message: ", err)
	}

	// Keep bot running until OS interruption
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	<-channel

	log.Println("\nShutting down gracefully.")
}
