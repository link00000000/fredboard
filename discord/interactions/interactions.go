package interactions

import (
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
)

var ErrOptNotFound = errors.New("option not found")
var ErrInvalidOptType = errors.New("invalid option type")

var ErrVoiceChannelNotFound = errors.New("voice channel not found")

// get an option from an interaction
//
// returns [ErrOptNotFound] when not found
func GetRequiredStringOpt(interaction *discordgo.Interaction, name string) (string, error) {
	var opt *discordgo.ApplicationCommandInteractionDataOption

	for _, o := range interaction.ApplicationCommandData().Options {
		if o.Name == name {
			opt = o
			break
		}
	}

	if opt == nil {
		return "", ErrOptNotFound
	}

	if opt.Type != discordgo.ApplicationCommandOptionString {
		return "", ErrInvalidOptType
	}

	return opt.StringValue(), nil
}

// search all voice channels of the guild that the interaction was created in for the interaction creator
func FindCreatorVoiceChannelId(session *discordgo.Session, interaction *discordgo.Interaction) (string, error) {
	if interaction.Type != discordgo.InteractionApplicationCommand && interaction.Type != discordgo.InteractionApplicationCommandAutocomplete {
		panic("joinVoiceChannelOfInteractionCreator called on interaction of type " + interaction.Type.String())
	}

	guild, err := session.State.Guild(interaction.GuildID)
	if err != nil {
		return "", err
	}

	var vc string
	for _, state := range guild.VoiceStates {
		if state.UserID == interaction.Member.User.ID {
			vc = state.ChannelID
			break
		}
	}

	if vc == "" {
		return "", ErrVoiceChannelNotFound
	}

	return vc, nil
}

// inform discord that the interaction has been acknowledged and will be responded to later
func Acknowledge(session *discordgo.Session, interaction *discordgo.Interaction) error {
	return session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

// respond to an interaction with a formatted error
func RespondWithError(session *discordgo.Session, interaction *discordgo.Interaction, message string, err error) error {
	return session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				&discordgo.MessageEmbed{Title: message, Description: err.Error(), Color: 0xeb3b40, Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")},
			},
		},
	})
}

// respond to an interaction with a plain string
func RespondWithMessage(session *discordgo.Session, interaction *discordgo.Interaction, message string) error {
	return session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: message},
	})
}
