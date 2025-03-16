package interactions

import (
	"errors"
	"time"

	"accidentallycoded.com/fredboard/v3/internal/audiosession"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/bwmarrin/discordgo"
)

var ErrOptNotFound = errors.New("option not found")
var ErrInvalidOptType = errors.New("invalid option type")

var ErrVoiceChannelNotFound = errors.New("discord voice channel not found")

var ErrVoiceConnectionNotFound = errors.New("discord voice connection does not exist")

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

func FindVoiceConn(session *discordgo.Session, interaction *discordgo.Interaction) (*discordgo.VoiceConnection, error) {
	if conn, ok := session.VoiceConnections[interaction.GuildID]; ok {
		return conn, nil
	}

	return nil, ErrVoiceConnectionNotFound
}

func FindOrCreateVoiceConn(session *discordgo.Session, interaction *discordgo.Interaction) (conn *discordgo.VoiceConnection, exists bool, err error) {
	if conn, ok := session.VoiceConnections[interaction.GuildID]; ok {
		return conn, true, nil
	}

	cId, err := FindCreatorVoiceChannelId(session, interaction)

	if err != nil {
		return nil, false, err
	}

	const mute = false
	const deaf = true
	conn, err = session.ChannelVoiceJoin(interaction.GuildID, cId, mute, deaf)

	return conn, false, err
}

func FindOrCreateAudioSession(logger *logging.Logger, session *discordgo.Session, interaction *discordgo.Interaction) (audioSession *audiosession.Session, output *audiosession.DiscordVoiceConnOutput, exists bool, err error) {
	conn, exists, err := FindOrCreateVoiceConn(session, interaction)
	if err != nil {
		return nil, nil, false, err
	}

	if exists {
		output, err = audiosession.FindDiscordVoiceConnOutput(conn)
		if err != nil {
			return nil, nil, true, err
		}

		return output.Session(), output, true, nil
	}

	audioSession = audiosession.New(logger)
	output, err = audioSession.AddDiscordVoiceConnOutput(conn)
	if err != nil {
		return nil, nil, false, err
		// TODO: destroy audio session
	}

	return audioSession, output, false, nil
}

// inform discord that the interaction has been acknowledged and will be responded to later
func Acknowledge(logger *logging.Logger, session *discordgo.Session, interaction *discordgo.Interaction) error {
	logger.Debug("responding to discord interaction with acknowledgment", "session", session, "interaction", interaction)

	err := Acknowledge_NoLog(session, interaction)
	if err != nil {
		logger.Error("failed to respond to discord interaction with acknowledgment", "session", session, "interaction", interaction, "error", err)
	}

	return err
}

// inform discord that the interaction has been acknowledged and will be responded to later
func Acknowledge_NoLog(session *discordgo.Session, interaction *discordgo.Interaction) error {
	err := session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	return err
}

// respond to an acknowledged interaction with a formatted error
func RespondWithError(logger *logging.Logger, session *discordgo.Session, interaction *discordgo.Interaction, inErr error) error {
	logger.Debug("responding to discord interaction with error", "session", session, "interaction", interaction, "inErr", inErr)

	err := RespondWithErrorMessage_NoLog(session, interaction, "There was an unexpected errror", inErr)
	if err != nil {
		logger.Error("failed to respond to discord interaction with error", "session", session, "interaction", interaction, "inErr", inErr, "error", err)
	}

	return inErr
}

// respond to an acknowledged interaction with a formatted error and a custom message
func RespondWithErrorMessage(logger *logging.Logger, session *discordgo.Session, interaction *discordgo.Interaction, message string, inErr error) error {
	logger.Debug("responding to discord interaction with error", "session", session, "interaction", interaction, "message", message, "inErr", inErr)

	err := RespondWithErrorMessage_NoLog(session, interaction, message, inErr)
	if err != nil {
		logger.Error("failed to respond to discord interaction with error", "session", session, "interaction", interaction, "message", message, "inErr", inErr, "error", err)
	}

	return inErr
}

// respond to an acknowledged interaction with a formatted error and a custom message
func RespondWithErrorMessage_NoLog(session *discordgo.Session, interaction *discordgo.Interaction, message string, inErr error) error {
	_, err := session.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{Title: message, Description: inErr.Error(), Color: 0xeb3b40, Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")},
		},
	})

	return err
}

// respond to an acknowledged interaction with a plain string
func RespondWithMessage(logger *logging.Logger, session *discordgo.Session, interaction *discordgo.Interaction, message string) error {
	logger.Debug("responding to discord interaction with message", "session", session, "interaction", interaction, "message", message)

	err := RespondWithMessage_NoLog(session, interaction, message)
	if err != nil {
		logger.Error("failed to respond to discord interaction with message", "session", session, "interaction", interaction, "message", message, "error", err)
	}

	return err
}

// respond to an acknowledged interaction with a plain string
func RespondWithMessage_NoLog(session *discordgo.Session, interaction *discordgo.Interaction, message string) error {
	_, err := session.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{
		Content: &message,
	})

	return err
}
