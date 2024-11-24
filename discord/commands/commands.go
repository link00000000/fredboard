package commands

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

var ErrNotFound = errors.New("not found")
var ErrInteractionCreatorNotInVoiceChannel = errors.New("interaction creator not in voice channel")

func getRequiredApplicationCommandOption(data discordgo.ApplicationCommandInteractionData, name string, optType discordgo.ApplicationCommandOptionType) (*discordgo.ApplicationCommandInteractionDataOption, error) {
	var foundOpt *discordgo.ApplicationCommandInteractionDataOption
	for _, opt := range data.Options {
		if opt.Name == name {
			foundOpt = opt
			break
		}
	}

	if foundOpt == nil {
		return nil, ErrNotFound
	}

	if foundOpt.Type != optType {
		return nil, ErrNotFound
	}

	return foundOpt, nil
}

func joinVoiceChannelIdOfInteractionCreator(session *discordgo.Session, interaction *discordgo.Interaction, mute, deaf bool) (*discordgo.VoiceConnection, error) {
	if interaction.Type != discordgo.InteractionApplicationCommand && interaction.Type != discordgo.InteractionApplicationCommandAutocomplete {
		panic("joinVoiceChannelOfInteractionCreator called on interaction of type " + interaction.Type.String())
	}

	guild, err := session.State.Guild(interaction.GuildID)

	if err != nil {
		return nil, err
	}

	var voiceChannelId string = ""
	for _, voiceState := range guild.VoiceStates {
		if voiceState.UserID == interaction.Member.User.ID {
			voiceChannelId = voiceState.ChannelID
			break
		}
	}

	if voiceChannelId == "" {
		return nil, ErrNotFound
	}

	voiceConnection, err := session.ChannelVoiceJoin(interaction.GuildID, voiceChannelId, mute, deaf)
	if err != nil {
		return nil, err
	}

	return voiceConnection, nil
}
