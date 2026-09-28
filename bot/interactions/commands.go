package interactions

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/royarchl/D2-server-bot/bot/database"
)

var dmPermission = false
var defaultMemberPermissions int64 = discordgo.PermissionManageGuild

var raidChoices = []*discordgo.ApplicationCommandOptionChoice{
	{Name: "Last Wish", Value: "Last Wish"},
	{Name: "Garden of Salvation", Value: "Garden of Salvation"},
	{Name: "Deep Stone Crypt", Value: "Deep Stone Crypt"},
	{Name: "Vault of Glass", Value: "Vault of Glass"},
	{Name: "Vow of the Disciple", Value: "Vow of the Disciple"},
	{Name: "King's Fall", Value: "King's Fall"},
	{Name: "Root of Nightmares", Value: "Root of Nightmares"},
	{Name: "Crota's End", Value: "Crota's End"},
	{Name: "Salvation's Edge", Value: "Salvation's Edge"},
	{Name: "The Desert Perpetual", Value: "The Desert Perpetual"},
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "roster",
		Description: "Manage raid rosters",
		Options: []*discordgo.ApplicationCommandOption{
			// Option 1: /roster create
			{
				Name:        "create",
				Description: "Create a new raid roster",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "raid",
						Description: "Choose a raid to create a roster for",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
						Choices:     raidChoices,
					},
					{
						Name:        "note",
						Description: "A brief note about your intent with this raid (optional)",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    false,
					},
				},
			},
			// Option 2: /roster edit
			{
				Name:        "edit",
				Description: "Edit an existing raid roster",
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "leader",
						Description: "Change the raid leader",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Name:        "raid",
								Description: "The raid to modify",
								Type:        discordgo.ApplicationCommandOptionString,
								Required:    true,
								Choices:     raidChoices,
							},
							{
								Name:        "user",
								Description: "The new leader",
								Type:        discordgo.ApplicationCommandOptionUser,
								Required:    true,
							},
						},
					},
					{
						Name:        "reserve",
						Description: "Choose a user to reserve a raid position for",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Name:        "raid",
								Description: "The raid to reserve the user for",
								Type:        discordgo.ApplicationCommandOptionString,
								Required:    true,
								Choices:     raidChoices,
							},
							{
								Name:        "user",
								Description: "The user to reserve",
								Type:        discordgo.ApplicationCommandOptionUser,
								Required:    true,
							},
						},
					},
				},
			},
			// Option 3: /roster notify
			{
				Name:        "notify",
				Description: "Notify roster members about raid starting",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "raid",
						Description: "Choose a raid to notify",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
						Choices:     raidChoices,
					},
					{
						Name:        "message",
						Description: "An announcement message (optional)",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    false,
					},
				},
			},
			// Option 4: /roster complete
			{
				Name:        "complete",
				Description: "Mark a raid as completed. (WARNING! DESTROYS ROSTER)",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "raid",
						Description: "Choose a raid to mark as completed",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
						Choices:     raidChoices,
					},
				},
			},
			// Option 5: /roster delete
			{
				Name:        "delete",
				Description: "Delete an empty roster. (WARNING! DESTROYS ROSTER)",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "raid",
						Description: "Choose a raid to mark as completed",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
						Choices:     raidChoices,
					},
				},
			},
		},
	},
	{
		Name:        "rotation",
		Description: "Query the weekly raid rotation",
		Type:        discordgo.ChatApplicationCommand,
	},
	{
		Name:        "guardian",
		Description: "GUARDIAN!",
		Type:        discordgo.ChatApplicationCommand,
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"roster":   handlerCommand_roster,
	"rotation": handlerCommand_rotation,
	"guardian": handlerCommand_guardian,
}

func parseCommandPath(options []*discordgo.ApplicationCommandInteractionDataOption) (string, []*discordgo.ApplicationCommandInteractionDataOption) {
	var pathSegments []string
	curr := options

	for len(curr) > 0 {
		opt := curr[0]
		if opt.Type == discordgo.ApplicationCommandOptionSubCommand ||
			opt.Type == discordgo.ApplicationCommandOptionSubCommandGroup {
			pathSegments = append(pathSegments, opt.Name)
			curr = opt.Options
		} else {
			break
		}
	}
	return strings.Join(pathSegments, " "), curr
}

func convertOptionsToMap(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	m := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, opt := range options {
		m[opt.Name] = opt
	}
	return m
}

func handlerCommand_roster(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	cmdPath, leafOptions := parseCommandPath(data.Options)
	optMap := convertOptionsToMap(leafOptions)

	// NOTE: The 'raid' options should return the name, not ID.
	switch cmdPath {
	case "create":
		handlerSubCommand_rosterCreate(s, i, optMap)
	case "edit leader":
		handlerSubCommand_rosterEditLeader(s, i, optMap)
	case "edit reserve":
		handlerSubCommand_rosterEditReserve(s, i, optMap)
	case "notify":
		handlerSubCommand_rosterNotify(s, i, optMap)
	case "complete":
		handlerSubCommand_rosterComplete(s, i, optMap)
	case "delete":
		handlerSubCommand_rosterDelete(s, i, optMap)
	}
}

func handlerSubCommand_rosterCreate(s *discordgo.Session, i *discordgo.InteractionCreate,
	options map[string]*discordgo.ApplicationCommandInteractionDataOption) {

	var response *string
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	defer func() {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: response,
		})
	}()

	raidName := options["raid"].StringValue()
	userID := i.Member.User.ID

	var note string
	if noteOpt, exists := options["note"]; exists {
		note = noteOpt.StringValue()
	}

	rosterID, err := database.RosterCreateDatabaseEntry(raidName, userID, note)
	if err != nil {
		log.Println(err)
		response = new("An active roster already exists for this raid.")
		return
	}

	// NOTE: Temporary channel reference
	channelID := os.Getenv("RAID_CHANNEL_ID")
	msg, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content:    generateRosterMessage(raidName),
		Components: RosterComponents,
	})

	database.RosterUpdateMessageId(msg.ID, rosterID)
	response = new(fmt.Sprintf("The roster (%s) has been created.", raidName))
}

func handlerSubCommand_rosterEditLeader(s *discordgo.Session, i *discordgo.InteractionCreate,
	options map[string]*discordgo.ApplicationCommandInteractionDataOption) {

	// NOTE: Verify the caller is the roster creator
	// NOTE: Creator should not be able to use on themselves.
	// NOTE: Creator should not be able to add users not on the roster.

	var response *string
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	defer func() {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: response,
		})
	}()

	raidName := options["raid"].StringValue()
	user := options["user"].UserValue(s)

	roster, err := database.RosterQueryRowFromRaidName(raidName)
	if err != nil {
		log.Println(err)
	}

	err = database.RosterUserUpsert(roster.ID, user.ID, database.StatusRaidLeader)

	channelID := os.Getenv("RAID_CHANNEL_ID")
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         *roster.MessageID,
		Channel:    channelID,
		Content:    new(generateRosterMessage(raidName)),
		Components: &RosterComponents,
	})

	response = new(fmt.Sprintf("The raid leader (%s) has been updated.", raidName))
}

func handlerSubCommand_rosterEditReserve(s *discordgo.Session, i *discordgo.InteractionCreate,
	options map[string]*discordgo.ApplicationCommandInteractionDataOption) {

	// NOTE: Creator should not be able to use on themselves.
	// NOTE: Creator should not be able to add users not on the roster.

	var response *string
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	defer func() {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: response,
		})
	}()

	raidName := options["raid"].StringValue()
	user := options["user"].UserValue(s)

	roster, err := database.RosterQueryRowFromRaidName(raidName)
	if err != nil {
		log.Println(err)
	}

	err = database.RosterUserUpsert(roster.ID, user.ID, database.StatusReserved)

	channelID := os.Getenv("RAID_CHANNEL_ID")
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         *roster.MessageID,
		Channel:    channelID,
		Content:    new(generateRosterMessage(raidName)),
		Components: &RosterComponents,
	})

	response = new(fmt.Sprintf("The user has been reserved (%s).", raidName))
}

func handlerSubCommand_rosterNotify(s *discordgo.Session, i *discordgo.InteractionCreate,
	options map[string]*discordgo.ApplicationCommandInteractionDataOption) {

	raidName := options["raid"].StringValue()

	// NOTE: Necessary since the 'message' property can be null.
	message := ""
	if opt, ok := options["message"]; ok && opt != nil {
		message = opt.StringValue()
	}

	roster, err := database.RosterQueryRowFromRaidName(raidName)
	if err != nil {
		log.Println(err)
	}
	users := database.RosterQueryUserIdsSorted(roster.ID)

	// NOTE: I want to ping the top 6 slots, but it's complicated since the list
	// is divided between two separate slices.
	var sb strings.Builder
	for _, user := range users.ActiveGroup {
		fmt.Fprintf(&sb, "<@%s> ", user.ID)
	}
	sb.WriteString(message)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: sb.String(),
		},
	})
	if err != nil {
		log.Println(err)
	}
}

func handlerSubCommand_rosterComplete(s *discordgo.Session, i *discordgo.InteractionCreate,
	options map[string]*discordgo.ApplicationCommandInteractionDataOption) {

	var response *string
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	defer func() {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: response,
		})
	}()

	raidName := options["raid"].StringValue()

	roster, err := database.RosterQueryRowFromRaidName(raidName)
	if err != nil {
		log.Println(err)
		response = new(fmt.Sprintf("No active roster exists for the **%s** raid.", raidName))
		return
	}

	if i.Member.User.ID != roster.AuthorID {
		response = new("You don't have permission to make this change. Only the **roster creator** can mark a raid as completed.")
		return
	}

	database.RosterArchive(roster.ID)

	// Set completions for users
	users := database.RosterQueryUserIdsSorted(roster.ID)
	const MAX_RAID_PARTICIPANTS = 6

	for _, user := range users.ActiveGroup {
		// NOTE: This should be one big transaction, instead of multiple writes.
		database.MarkUserRaidCompletion(user.ID, raidName)
	}

	channelID := os.Getenv("RAID_CHANNEL_ID")
	s.ChannelMessageDelete(channelID, *roster.MessageID)

	response = new(fmt.Sprintf("The roster (%s) has been archived.", raidName))
}

func handlerSubCommand_rosterDelete(s *discordgo.Session, i *discordgo.InteractionCreate,
	options map[string]*discordgo.ApplicationCommandInteractionDataOption) {

	var response *string
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	defer func() {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: response,
		})
	}()

	raidName := options["raid"].StringValue()

	roster, err := database.RosterQueryRowFromRaidName(raidName)
	if err != nil {
		log.Println(err)
		response = new(fmt.Sprintf("No active roster exists for the **%s** raid.", raidName))
		return
	}

	if i.Member.User.ID != roster.AuthorID {
		response = new("You don't have permission to make this change. Only the **roster creator** can mark a raid as completed.")
		return
	}

	// NOTE: We aren't checking if there are additional users.

	database.RosterArchive(roster.ID)

	channelID := os.Getenv("RAID_CHANNEL_ID")
	s.ChannelMessageDelete(channelID, *roster.MessageID)

	response = new(fmt.Sprintf("The roster (%s) has been deleted.", raidName))
}

func handlerCommand_guardian(s *discordgo.Session, i *discordgo.InteractionCreate) {
	response := "" +
		"`░██████╗░██╗░░░██╗░█████╗░██████╗░██████╗░██╗░█████╗░███╗░░██╗██╗`\n" +
		"`██╔════╝░██║░░░██║██╔══██╗██╔══██╗██╔══██╗██║██╔══██╗████╗░██║██║`\n" +
		"`██║░░██╗░██║░░░██║███████║██████╔╝██║░░██║██║███████║██╔██╗██║██║`\n" +
		"`██║░░╚██╗██║░░░██║██╔══██║██╔══██╗██║░░██║██║██╔══██║██║╚████║╚═╝`\n" +
		"`╚██████╔╝╚██████╔╝██║░░██║██║░░██║██████╔╝██║██║░░██║██║░╚███║██╗`\n" +
		"`░╚═════╝░░╚═════╝░╚═╝░░╚═╝╚═╝░░╚═╝╚═════╝░╚═╝╚═╝░░╚═╝╚═╝░░╚══╝╚═╝`"

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
	if err != nil {
		panic(err)
	}
}

func handlerCommand_rotation(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// NOTE: This has yet to be implemented.

	fmt.Println("rotation")
}

// var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate,
// 	args []*discordgo.ApplicationCommandInteractionDataOption){
//
// 	"roster create": func(s *discordgo.Session, i *discordgo.InteractionCreate,
// 		args []*discordgo.ApplicationCommandInteractionDataOption) {
//
// 		// raidId := args[0].StringValue()
// 		// response := fmt.Sprintf("Creating roster for raid ID: %s", raidId)
// 	},
//
// 	"roster view": func(s *discordgo.Session, i *discordgo.InteractionCreate,
// 		args []*discordgo.ApplicationCommandInteractionDataOption) {
//
// 		// raidId := args[0].StringValue()
// 		// response := fmt.Sprintf("Creating roster for raid ID: %s", raidId)
// 	},
//
// 	"roster edit leader": func(s *discordgo.Session, i *discordgo.InteractionCreate,
// 		args []*discordgo.ApplicationCommandInteractionDataOption) {
//
// 		// raidId := args[0].StringValue()
// 		// response := fmt.Sprintf("Creating roster for raid ID: %s", raidId)
// 	},
//
// 	"roster edit reserve": func(s *discordgo.Session, i *discordgo.InteractionCreate,
// 		args []*discordgo.ApplicationCommandInteractionDataOption) {
//
// 		// raidId := args[0].StringValue()
// 		// response := fmt.Sprintf("Creating roster for raid ID: %s", raidId)
// 	},
//
// 	"roster notify": func(s *discordgo.Session, i *discordgo.InteractionCreate,
// 		args []*discordgo.ApplicationCommandInteractionDataOption) {
//
// 		// raidId := args[0].StringValue()
// 		// response := fmt.Sprintf("Creating roster for raid ID: %s", raidId)
// 	},
//
// 	"roster complete": func(s *discordgo.Session, i *discordgo.InteractionCreate,
// 		args []*discordgo.ApplicationCommandInteractionDataOption) {
//
// 		// raidId := args[0].StringValue()
// 		// response := fmt.Sprintf("Creating roster for raid ID: %s", raidId)
// 	},
//
// 	"rotation": func(s *discordgo.Session, i *discordgo.InteractionCreate,
// 		args []*discordgo.ApplicationCommandInteractionDataOption) {
//
// 		// NOTE: This information needs to go in the function that sets the
// 		// IN_ROTATION value in the database for the raids.
// 		_raid := map[string]int{
// 			"Last Wish":            1,
// 			"Garden of Salvation":  0,
// 			"Deep Stone Crypt":     0,
// 			"Vault of Glass":       0,
// 			"Vow of the Disciples": 1,
// 			"King's Fall":          0,
// 			"Root of Nightmares":   0,
// 			"Crota's End":          0,
// 			"Salvation's Edge":     0,
// 		}
//
// 		raid := map[string]bool{
// 			"Last Wish":            true,
// 			"Garden of Salvation":  false,
// 			"Deep Stone Crypt":     false,
// 			"Vault of Glass":       false,
// 			"Vow of the Disciples": true,
// 			"King's Fall":          false,
// 			"Root of Nightmares":   false,
// 			"Crota's End":          false,
// 			"Salvation's Edge":     false,
// 		}
//
// 		// Asterisk on what’s featured this week. Move them both down each week.
// 		raids := [9]string{
// 			"Last Wish", // *
// 			"Garden of Salvation",
// 			"Deep Stone Crypt",
// 			"Vault of Glass",
// 			"Vow of the Disciples", // *
// 			"King's Fall",
// 			"Root of Nightmares",
// 			"Crota's End",
// 			"Salvation's Edge",
// 		}
//
// 		fmt.Print(_raid)
// 		fmt.Print(raid)
// 		fmt.Print(raids)
//
// 		// raidId := args[0].StringValue()
// 		// response := fmt.Sprintf("Creating roster for raid ID: %s", raidId)
// 	},
//
// 	"guardian": func(s *discordgo.Session, i *discordgo.InteractionCreate,
// 		args []*discordgo.ApplicationCommandInteractionDataOption) {
//
// response := "" +
// 	"`░██████╗░██╗░░░██╗░█████╗░██████╗░██████╗░██╗░█████╗░███╗░░██╗██╗`\n" +
// 	"`██╔════╝░██║░░░██║██╔══██╗██╔══██╗██╔══██╗██║██╔══██╗████╗░██║██║`\n" +
// 	"`██║░░██╗░██║░░░██║███████║██████╔╝██║░░██║██║███████║██╔██╗██║██║`\n" +
// 	"`██║░░╚██╗██║░░░██║██╔══██║██╔══██╗██║░░██║██║██╔══██║██║╚████║╚═╝`\n" +
// 	"`╚██████╔╝╚██████╔╝██║░░██║██║░░██║██████╔╝██║██║░░██║██║░╚███║██╗`\n" +
// 	"`░╚═════╝░░╚═════╝░╚═╝░░╚═╝╚═╝░░╚═╝╚═════╝░╚═╝╚═╝░░╚═╝╚═╝░░╚══╝╚═╝`"
//
// 		// response := strings.ReplaceAll(rawStr, "§", "`")
//
// 		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
// 			Type: discordgo.InteractionResponseChannelMessageWithSource,
// 			Data: &discordgo.InteractionResponseData{
// 				Content: response,
// 			},
// 		})
// 	},
// }

// func setCommands(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
// 	if interaction.Type != discordgo.InteractionApplicationCommand {
// 		return
// 	}
//
// 	data := interaction.ApplicationCommandData()
// 	path, args := getCommandPath(data)
//
// 	if handler, ok := commandHandlers[path]; ok {
// 		handler(session, interaction, args)
// 	} else {
// 		log.Printf("No handler found for command path: %s", path)
// 	}
//
// 	// if handler, ok := commandHandlers[interaction.ApplicationCommandData().Name]; ok {
// 	// 	handler(session, interaction)
// 	// }
// }

// func registerCommands(session *discordgo.Session) {
// 	for _, v := range commands {
// 		guildID := os.Getenv("GUILD_ID")
// 		_, err := session.ApplicationCommandCreate(session.State.User.ID, guildID, v)
// 		if err != nil {
// 			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
// 		}
// 	}
// }

// Broken
// func removeCommands(session *discordgo.Session) {
// 	for _, cmd := range commands {
// 		guildID := os.Getenv("GUILD_ID")
// 		err := session.ApplicationCommandDelete(session.State.User.ID, guildID, cmd.ID)
// 		if err != nil {
// 			log.Panicf("Cannot delete '%v' command: %v", cmd.Name, err)
// 		}
// 	}
// }
