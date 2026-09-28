package interactions

import (
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/royarchl/D2-server-bot/bot/database"
)

// This is where the button interactions will be handled.

type RoleButton struct {
	Name   string
	RoleID string
}

// var roles []RoleButton
// var RoleComponents []discordgo.MessageComponent

var RaidRoles = map[string]bool{
	"Forsaken":         true,
	"Shadowkeep":       true,
	"Beyond Light":     true,
	"30th Anniversary": true,
	"Witch Queen":      true,
	"Lightfall":        true,
	"Final Shape":      true,
	"Edge of Fate":     true,
	"Renegades":        true,
	"Dungeon Pack":     true,
}

var TimezoneRoles = map[string]bool{
	"UTC-8 (PST)": true,
	"UTC-6 (CST)": true,
	"UTC-5 (EST)": true,
	"UTC+0 (GMT)": true,
}

var ExperienceRoles = map[string]bool{
	"Veteran":    true,
	"New Player": true,
}

// NOTE: These will need to be called on startup, when generating the role messages.

func GetFilteredRoleList(s *discordgo.Session, accepted_roles map[string]bool) ([]RoleButton, error) {
	guildID := os.Getenv("GUILD_ID")

	guildRoles, err := s.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}

	var filteredList []RoleButton

	for _, role := range guildRoles {
		if accepted_roles[role.Name] {
			filteredList = append(filteredList, RoleButton{
				Name:   role.Name,
				RoleID: role.ID,
			})
		}
	}

	return filteredList, nil
}

func GenerateRoleComponents(roles []RoleButton) []discordgo.MessageComponent {
	var rows []discordgo.MessageComponent
	var currentRow []discordgo.MessageComponent

	for i, role_button := range roles {
		btn := discordgo.Button{
			Label:    role_button.Name,
			Style:    discordgo.SecondaryButton,
			CustomID: "role_button_" + role_button.RoleID,
		}
		currentRow = append(currentRow, btn)

		// Discord allows max 5 buttons per row, else it's the last item
		if len(currentRow) == 5 || i == len(roles)-1 {
			rows = append(rows, discordgo.ActionsRow{Components: currentRow})
			currentRow = []discordgo.MessageComponent{}
		}
	}
	return rows
}

// NOTE:
// This will be private (RosterComponents → rosterComponents) after the
// `/roster create` command is ported to the `commands.go` file.
var RosterComponents = []discordgo.MessageComponent{
	discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "✔️ Interested",
				Style:    discordgo.SecondaryButton,
				CustomID: "roster_join",
			},
			discordgo.Button{
				Label:    "⏳ Maybe",
				Style:    discordgo.SecondaryButton,
				CustomID: "roster_maybe",
			},
			discordgo.Button{
				Label:    "❌ Not Interested",
				Style:    discordgo.SecondaryButton,
				CustomID: "roster_decline",
			},
		},
	},
}

var componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"roster_join":    handlerButton_rosterJoin,
	"roster_maybe":   handlerButton_rosterMaybe,
	"roster_decline": handlerButton_rosterDecline,
}

func handlerButton_rosterJoin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// NOTE:
	// Check if the user exists in the roster.
	// If yes, simply return.
	// If no, add them to the roster.
	// Update the roster message to reflect their participation.
	// - Will most likely just require re-building the entire message string,
	//   unless you think of something more efficient.
	// NOTE: The message should only be updated if any changes were made.

	// NOTE: This is correct form, which changes the original message.
	// msg := fmt.Sprintf("%s JOIN", i.Interaction.Message.Content)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		panic(err)
	}

	// NOTE: I should be able to check if they already have 'StatusStandby' set.

	rosterID, err := database.RosterQueryIdFromMessageId(i.Interaction.Message.ID)
	err = database.RosterUserUpsert(rosterID, i.Interaction.Member.User.ID, database.StatusActive)
	raidName, err := database.RosterQueryRaidNameFromId(rosterID)

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: new(generateRosterMessage(raidName)),
	})
}

func handlerButton_rosterMaybe(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// NOTE:
	// Check if the user exists in the roster.
	// If yes, ...
	// 		If active roster, move to standby
	//		If on standby, simply return.
	// If no, add them to standby section of roster.
	// Update the roster message to reflect their participation.
	// - Will most likely just require re-building the entire message string,
	//   unless you think of something more efficient.
	// NOTE: The message should only be updated if any changes were made.
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		panic(err)
	}

	// NOTE: I should be able to check if they already have 'StatusStandby' set.
	rosterID, err := database.RosterQueryIdFromMessageId(i.Interaction.Message.ID)
	err = database.RosterUserUpsert(rosterID, i.Interaction.Member.User.ID, database.StatusStandby)
	raidName, err := database.RosterQueryRaidNameFromId(rosterID)

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: new(generateRosterMessage(raidName)),
	})
}

func handlerButton_rosterDecline(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// NOTE:
	// Check if the user exists in the roster.
	// If yes, remove entry from roster table.
	// If no, simply return.
	// Update the roster message to reflect their participation.
	// - Will most likely just require re-building the entire message string,
	//   unless you think of something more efficient.
	// NOTE: The message should only be updated if any changes were made.
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		panic(err)
	}

	rosterID, err := database.RosterQueryIdFromMessageId(i.Interaction.Message.ID)
	err = database.RosterUserLeave(rosterID, i.Interaction.Member.User.ID)
	if err != nil {
		log.Println(err)
		return
	}
	raidName, err := database.RosterQueryRaidNameFromId(rosterID)

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: new(generateRosterMessage(raidName)),
	})
}

func handlerButton_roleAssign(s *discordgo.Session, i *discordgo.InteractionCreate) {
	roleID := strings.TrimPrefix(i.MessageComponentData().CustomID, "role_button_")
	userID := i.Member.User.ID

	err := s.GuildMemberRoleAdd(i.GuildID, userID, roleID)
	if err != nil {
		panic(err)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		panic(err)
	}

	// Message feedback
}
