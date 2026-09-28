package interactions

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/royarchl/D2-server-bot/bot/database"
)

var guildCommandIds []string
var GuildId string
var RolesChannelId string

func RegisterInteractions(s *discordgo.Session) {
	registerHandlers(s)
	registerCommands(s)
}

func DeregisterCommands(s *discordgo.Session) {
	commands, err := s.ApplicationCommands(s.State.User.ID, GuildId)
	if err != nil {
		log.Fatal(err)
	}

	for _, cmd := range commands {
		err := s.ApplicationCommandDelete(s.State.User.ID, GuildId, cmd.ID)
		if err != nil {
			log.Panic(err)
		}
	}
}

func registerHandlers(s *discordgo.Session) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
			customID := i.MessageComponentData().CustomID

			if h, ok := componentHandlers[customID]; ok {
				h(s, i)
				return
			}

			// Dynamix prefix matches
			if strings.HasPrefix(customID, "role_button_") {
				handlerButton_roleAssign(s, i)
				return
			}
		}
	})
}

func registerCommands(s *discordgo.Session) {
	for _, cmd := range commands {
		c, err := s.ApplicationCommandCreate(s.State.User.ID, GuildId, cmd)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", cmd.Name, err)
		}
		guildCommandIds = append(guildCommandIds, c.ID)
	}
}

// NOTE: Create the function that would generate the reaction buttons on startup

func generateRosterMessage(raid_name string) string {
	const rosterAuthorIcon string = "`📋`"
	const rosterLeaderIcon string = "`📣`"
	const rosterNoClearsIcon string = "`☘️`"
	const rosterReserved string = "`📌`"
	const raidInRotationIcon string = "`🔄 In-Rotation`"

	const MAX_RAID_PARTICIPANTS = 6

	var sb strings.Builder
	raidDbEntry, err := database.RaidQueryRowFromName(raid_name)
	if err != nil {
		log.Println(err)
		return ""
	}

	roster, err := database.RosterQueryRowFromRaidName(raidDbEntry.Name)

	// Begin message generation
	fmt.Fprintf(&sb, "Guardians,\n-# _ _\nWe are looking for people with the <@&%s> DLC to complete a raid.\n", raidDbEntry.RoleID)

	if *roster.Note != "" {
		fmt.Fprintf(&sb, "_ _\n> _Author's Note_\n> %s\n", *roster.Note)
	}

	fmt.Fprintf(&sb, "-# _ _\n### RAID:  **%s**  ", raidDbEntry.Name)

	if raidDbEntry.InRotation {
		sb.WriteString(raidInRotationIcon)
	}
	sb.WriteString("\n")

	fmt.Fprintf(&sb, "_%s_\n-# _ _\n", raidDbEntry.Description)

	// Printing participating users
	users := database.RosterQueryUserIdsSorted(roster.ID)

	applyExtraTag := func(status int64) string {
		switch status {
		case database.StatusCreator:
			return rosterAuthorIcon
		case database.StatusRaidLeader:
			return rosterLeaderIcon
		case database.StatusNoClears:
			return rosterNoClearsIcon
		case database.StatusReserved:
			return rosterReserved
		default:
			return ""
		}
	}

	sb.WriteString("### -# _Participating Members_\n")
	for i, user := range users.ActiveGroup {
		fmt.Fprintf(&sb, "%d. <@%s> %s\n", i, user.ID, applyExtraTag(user.Status))
	}

	sb.WriteString("### -# _On Standby_\n")
	for _, user := range users.StandbyGroup {
		fmt.Fprintf(&sb, "- <@%s> %s\n", user.ID, applyExtraTag(user.Status))
	}

	sb.WriteString("-# _ _\n-# _ _\n**Select an option:** ")

	return sb.String()
}

func GenerateRolesMessage(s *discordgo.Session, roles map[string]bool, content string) error {
	roleList, err := GetFilteredRoleList(s, roles)
	if err != nil {
		log.Println(err)
		return err
	}
	buttonComponents := GenerateRoleComponents(roleList)

	_, err = s.ChannelMessageSendComplex(RolesChannelId, &discordgo.MessageSend{
		Content:    content,
		Components: buttonComponents,
	})
	return nil
}
