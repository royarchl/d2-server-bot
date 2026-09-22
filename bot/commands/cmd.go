package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

var dmPermission = false
var defaultMemberPermissions int64 = discordgo.PermissionManageGuild

var raidChoices = []*discordgo.ApplicationCommandOptionChoice{
	{Name: "Last Wish", Value: "0000000000000000001"},
	{Name: "Garden of Salvation", Value: "0000000000000000002"},
	{Name: "Deep Stone Crypt", Value: "0000000000000000003"},
	{Name: "Vault of Glass", Value: "0000000000000000004"},
	{Name: "Vow of the Disciple", Value: "0000000000000000005"},
	{Name: "King's Fall", Value: "0000000000000000006"},
	{Name: "Root of Nightmare", Value: "0000000000000000007"},
	{Name: "Crota's End", Value: "0000000000000000008"},
	{Name: "Salvation's Edge", Value: "0000000000000000009"},
	{Name: "The Desert Perpetual", Value: "0000000000000000010"},
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
				},
			},
			// Option 2: /roster view
			{
				Name:        "view",
				Description: "View an existing raid roster",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "raid",
						Description: "Choose a raid to create a roster for",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
						Choices:     raidChoices,
					},
				},
			},
			// Option 3: /roster edit ...
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
			// Option 4: /roster notify
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
			// Option 5: /roster complete
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

func getCommandPath(data discordgo.ApplicationCommandInteractionData) (string, []*discordgo.ApplicationCommandInteractionDataOption) {
	path := data.Name
	commandOptions := data.Options

	for len(commandOptions) > 0 && (commandOptions[0].Type == discordgo.ApplicationCommandOptionSubCommandGroup || commandOptions[0].Type == discordgo.ApplicationCommandOptionSubCommand) {
		path += " " + commandOptions[0].Name
		commandOptions = commandOptions[0].Options
	}

	for len(commandOptions) > 0 {
		currentOption := commandOptions[0]

		isSubCommand := currentOption.Type == discordgo.ApplicationCommandOptionSubCommandGroup ||
			currentOption.Type == discordgo.ApplicationCommandOptionSubCommand

		if !isSubCommand {
			break
		}

		path += " " + currentOption.Name
		commandOptions = currentOption.Options
	}

	return path, commandOptions
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate,
	args []*discordgo.ApplicationCommandInteractionDataOption){

	"roster create": func(s *discordgo.Session, i *discordgo.InteractionCreate,
		args []*discordgo.ApplicationCommandInteractionDataOption) {

		// raidId := args[0].StringValue()
		// response := fmt.Sprintf("Creating roster for raid ID: %s", raidId)
	},

	"roster view": func(s *discordgo.Session, i *discordgo.InteractionCreate,
		args []*discordgo.ApplicationCommandInteractionDataOption) {

		// raidId := args[0].StringValue()
		// response := fmt.Sprintf("Creating roster for raid ID: %s", raidId)
	},

	"roster edit leader": func(s *discordgo.Session, i *discordgo.InteractionCreate,
		args []*discordgo.ApplicationCommandInteractionDataOption) {

		// raidId := args[0].StringValue()
		// response := fmt.Sprintf("Creating roster for raid ID: %s", raidId)
	},

	"roster edit reserve": func(s *discordgo.Session, i *discordgo.InteractionCreate,
		args []*discordgo.ApplicationCommandInteractionDataOption) {

		// raidId := args[0].StringValue()
		// response := fmt.Sprintf("Creating roster for raid ID: %s", raidId)
	},

	"roster notify": func(s *discordgo.Session, i *discordgo.InteractionCreate,
		args []*discordgo.ApplicationCommandInteractionDataOption) {

		// raidId := args[0].StringValue()
		// response := fmt.Sprintf("Creating roster for raid ID: %s", raidId)
	},

	"roster complete": func(s *discordgo.Session, i *discordgo.InteractionCreate,
		args []*discordgo.ApplicationCommandInteractionDataOption) {

		// raidId := args[0].StringValue()
		// response := fmt.Sprintf("Creating roster for raid ID: %s", raidId)
	},

	"rotation": func(s *discordgo.Session, i *discordgo.InteractionCreate,
		args []*discordgo.ApplicationCommandInteractionDataOption) {

		// NOTE: This information needs to go in the function that sets the
		// IN_ROTATION value in the database for the raids.
		_raid := map[string]int{
			"Last Wish":            1,
			"Garden of Salvation":  0,
			"Deep Stone Crypt":     0,
			"Vault of Glass":       0,
			"Vow of the Disciples": 1,
			"King's Fall":          0,
			"Root of Nightmares":   0,
			"Crota's End":          0,
			"Salvation's Edge":     0,
		}

		raid := map[string]bool{
			"Last Wish":            true,
			"Garden of Salvation":  false,
			"Deep Stone Crypt":     false,
			"Vault of Glass":       false,
			"Vow of the Disciples": true,
			"King's Fall":          false,
			"Root of Nightmares":   false,
			"Crota's End":          false,
			"Salvation's Edge":     false,
		}

		// Asterisk on what’s featured this week. Move them both down each week.
		raids := [9]string{
			"Last Wish", // *
			"Garden of Salvation",
			"Deep Stone Crypt",
			"Vault of Glass",
			"Vow of the Disciples", // *
			"King's Fall",
			"Root of Nightmares",
			"Crota's End",
			"Salvation's Edge",
		}

		fmt.Print(_raid)
		fmt.Print(raid)
		fmt.Print(raids)

		// raidId := args[0].StringValue()
		// response := fmt.Sprintf("Creating roster for raid ID: %s", raidId)
	},

	"guardian": func(s *discordgo.Session, i *discordgo.InteractionCreate,
		args []*discordgo.ApplicationCommandInteractionDataOption) {

		response := "" +
			"`░██████╗░██╗░░░██╗░█████╗░██████╗░██████╗░██╗░█████╗░███╗░░██╗██╗`\n" +
			"`██╔════╝░██║░░░██║██╔══██╗██╔══██╗██╔══██╗██║██╔══██╗████╗░██║██║`\n" +
			"`██║░░██╗░██║░░░██║███████║██████╔╝██║░░██║██║███████║██╔██╗██║██║`\n" +
			"`██║░░╚██╗██║░░░██║██╔══██║██╔══██╗██║░░██║██║██╔══██║██║╚████║╚═╝`\n" +
			"`╚██████╔╝╚██████╔╝██║░░██║██║░░██║██████╔╝██║██║░░██║██║░╚███║██╗`\n" +
			"`░╚═════╝░░╚═════╝░╚═╝░░╚═╝╚═╝░░╚═╝╚═════╝░╚═╝╚═╝░░╚═╝╚═╝░░╚══╝╚═╝`"

		// response := strings.ReplaceAll(rawStr, "§", "`")

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	},
}

func SetCommands(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := interaction.ApplicationCommandData()
	path, args := getCommandPath(data)

	if handler, ok := commandHandlers[path]; ok {
		handler(session, interaction, args)
	} else {
		log.Printf("No handler found for command path: %s", path)
	}

	// if handler, ok := commandHandlers[interaction.ApplicationCommandData().Name]; ok {
	// 	handler(session, interaction)
	// }
}

func RegisterCommands(session *discordgo.Session) {
	for _, v := range commands {
		guildID := os.Getenv("GUILD_ID")
		_, err := session.ApplicationCommandCreate(session.State.User.ID, guildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
	}
}
