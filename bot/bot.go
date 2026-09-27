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

	// Register commands and buttons (+handlers)
	interactions.GuildId = os.Getenv("GUILD_ID")
	interactions.RegisterInteractions(session)
	defer interactions.DeregisterCommands(session)

	// Connect database and register users
	err = database.EstablishDatabaseConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Db.Close()

	guildMembers, err := session.GuildMembers(interactions.GuildId, "", 1000)
	database.RegisterServerMembers(guildMembers)

	// guildRoles, err := session.GuildRoles(interactions.GuildId)
	// for _, role := range guildRoles {
	// 	if role.Name == "@everyone" {
	// 		continue
	// 	}

	// 	fmt.Printf("NAME: %s; ID: %s\n", role.Name, role.ID)
	// }

	// Send the "bot online" ack message
	message_botOnline := "*I'm back! Oh, my head... It feels like I was squeezed into a tiny little box.*"
	_, err = session.ChannelMessageSend(BotChannelId, message_botOnline)
	if err != nil {
		log.Println("Failed to send startup message: ", err)
	}

	// TEMPORARY - GENERATING A ROSTER
	// rosterId, err := database.RosterCreateDatabaseEntry("Last Wish", os.Getenv("TEST_USER_1_ID"))
	// if err != nil {
	// log.Println("There was an error creating the roster: ", err)
	// }
	//
	// rawStr := `Guardians,
	// -# _ _
	// We are looking for people with the <@&RAID_ROLE_ID> DLC to complete a raid.
	// _ _
	// > _Author's Note_
	// > I wanted to play this mission on Master difficulty to get exotic gear.
	// -# _ _
	// ### RAID:  Garden of Salvation  §🔄 In-Rotation§
	// _When Vex from the Sol Divisive are drawn to the awakened Pyramid beneath the Moon, Eris Morn tracks a mysterious signal with the Unknown Artifact emitting from deep within the Black Garden. Now Guardians must enter the birthplace of the Vex as they trace the signal to its source at the center of the garden, uncover remnants of the Black Fleet and defeat the Vex Minds empowered by them._
	// -# _ _
	// -# Based on current participants, the ideal time to begin would be <t:1789077600:t>.
	// -# _ _
	// ### -# _Participating Members_  _  _ [_ʜᴏᴡ ɪs ᴘʀɪᴏʀɪᴛʏ ᴅᴇᴄɪᴅᴇᴅ?_](<https://t>)
	// 1. <@TEST_USER_2_ID> §📋§ §📣§
	// -# (roster creator) (raid leader)
	// 2. <@TEST_USER_1_ID> §☘️§
	// -# (0 completions)
	// 3. <@TEST_USER_2_ID>
	// 4. <@TEST_USER_1_ID>
	// 5.
	// 6.
	// ### -# _On Standby_
	// - <@TEST_USER_2_ID>
	// -# _ _
	// -# _ _
	// **Select an option:**`
	//
	// replacer := strings.NewReplacer(
	// "§", "`",
	// "RAID_ROLE_ID", os.Getenv("RAID_ROLE_ID"),
	// "TEST_USER_1_ID", os.Getenv("TEST_USER_1_ID"),
	// "TEST_USER_2_ID", os.Getenv("TEST_USER_2_ID"),
	// )
	// // message := strings.ReplaceAll(rawStr, "§", "`")
	// message := replacer.Replace(rawStr)
	//
	// msg, err := session.ChannelMessageSendComplex(BotChannelId, &discordgo.MessageSend{
	// Content:    message,
	// Components: interactions.RosterComponents,
	// })
	//
	// if err != nil {
	// log.Println("Error sending message: ", err)
	// }
	//
	// err = database.RosterUpdateMessageId(msg.ID, rosterId)
	// if err != nil {
	// log.Println("Error updating message id: ", err)
	// }

	// const TEST_VALUE = 6
	// for i := range TEST_VALUE {
	// fmt.Println(i)
	// }

	// Keep bot running until OS interruption
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	<-channel

	log.Println("\nShutting down gracefully.")
}
