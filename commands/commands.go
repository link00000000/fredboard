package commands

import (
	"errors"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

var ErrNotFound = errors.New("not found")
var ErrInteractionCreatorNotInVoiceChannel = errors.New("interaction creator not in voice channel")

var logger = slog.Default()

func getRequiredApplicationCommandOption(data discordgo.ApplicationCommandInteractionData, name string, optType discordgo.ApplicationCommandOptionType) (*discordgo.ApplicationCommandInteractionDataOption, error) {
  logger.Debug("Getting required option", "data", data, "name", name, "optType", optType)  

  var foundOpt *discordgo.ApplicationCommandInteractionDataOption
  for _, opt := range data.Options {
    if opt.Name == name {
      foundOpt = opt
      break
    }
  }

  if foundOpt == nil {
    logger.Debug("Did not find required option", "data", data, "name", name, "optType", optType)
    return nil, ErrNotFound
  }

  if foundOpt.Type != optType {
    logger.Debug("Found option is not the correct type", "data", data, "name", name, "optType", optType)
    return nil, ErrNotFound
  }

  return foundOpt, nil
}

func findVoiceChannelIdOfInteractionCreator(session *discordgo.Session, interaction *discordgo.Interaction) (string, error) {
	if interaction.Type != discordgo.InteractionApplicationCommand && interaction.Type != discordgo.InteractionApplicationCommandAutocomplete {
		panic("joinVoiceChannelOfInteractionCreator called on interaction of type " + interaction.Type.String())
	}

  guild, err := session.State.Guild(interaction.GuildID)
  
  if err != nil {
    logger.Error("session.State.Guild()", "error", err, "session", session, "interaction", interaction)
    return "", err
  } else {
    logger.Debug("session.State.Guild()", "session", session, "interaction", interaction, "guild", guild)
  }

  var voiceChannelId string = ""
  for _, voiceState := range guild.VoiceStates {
    if voiceState.UserID == interaction.Member.User.ID {
      voiceChannelId = voiceState.ChannelID
      break
    }
  }

  if voiceChannelId == "" {
    logger.Debug("Failed to find voice channel for interaction creator", "session", session, "interaction", interaction, "guild", guild)
    return "", ErrNotFound
  } else {
    logger.Debug("Found voice channel for interaction creator", "session", session, "interaction", interaction, "guild", guild, "voiceChannelId", voiceChannelId)
  }

  return voiceChannelId, nil
}
