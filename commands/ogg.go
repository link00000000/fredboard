package commands

import (
	"fmt"
	"io"
	"os"
	"time"

	"accidentallycoded.com/fredboard/v3/codecs"
	"github.com/bwmarrin/discordgo"
)

func Ogg(session *discordgo.Session, interaction *discordgo.Interaction) (*discordgo.InteractionResponse, error) {
  interactionData := interaction.ApplicationCommandData()

  path, err := getRequiredApplicationCommandOption(interactionData, "path", discordgo.ApplicationCommandOptionString)
  if err != nil {
    logger.Error("Ogg: Failed to get required application option", "session", session, "interaction", interaction, "name", "path", "error", err)
    return nil, err
  }

  go playOggFile(session, interaction, path.StringValue())

  response := &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
  }

  return response, nil
}

func playOggFile(session *discordgo.Session, interaction *discordgo.Interaction, path string) {
  f, err := os.Open(path)
  if err != nil {
    logger.Error("playOggFile: Failed to open file", "path", path, "error", err)
    // TODO: Notify the user that there was an error
    return
  }

  logger.Debug("playOggFile: Opened file", "path", path, "file", f)

  defer func() {
    err := f.Close()

    if err != nil {
      logger.Error("playOggFile: Failed to close file", "file", f)
      return
    }

    logger.Debug("playOggFile: Closed file", "file", f)
  }()

  voiceChannelId, err := findVoiceChannelIdOfInteractionCreator(session, interaction)
  if err != nil {
    logger.Error("playOggFile: Failed to find join voice channel of interaction creator", "session", session, "interaction", interaction, "error", err)
    // TODO: Notify the user that there was an error
    return
  }
  
  logger.Debug("playOggFile: Found voice channel of interaction creator", "session", session, "interaction", interaction, "voiceChannelId", voiceChannelId)

  const mute = false
  const deaf = true
  voiceConnection, err := session.ChannelVoiceJoin(interaction.GuildID, voiceChannelId, mute, deaf)

  if err != nil {
    logger.Error("playOggFile: Failed to join voice channel", "session", session, "interaction", interaction, "voiceChannelId", voiceChannelId)
    // TODO: Notify the user that there was an error
    return
  }

  defer func() {
    err := voiceConnection.Disconnect()

    if err != nil {
      logger.Error("playOggFile: Failed to close voice connection", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
      return
    }

    logger.Debug("playOggFile: Closed voice channel", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
  }()

  time.Sleep(250 * time.Millisecond) // Give voice connection time to settle

  session.InteractionRespond(interaction, &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseChannelMessageWithSource,
    Data: &discordgo.InteractionResponseData{
      Content: fmt.Sprintf("Playing %s", f.Name),
    },
  })

  reader := codecs.NewOggOpusReader(f)

  err = voiceConnection.Speaking(true)
  if err != nil {
    logger.Error("playOggFile: Failed to set speaking status to true", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
    // TODO: Notify the user that there was an error
    return
  }

  logger.Debug("playOggFile: Set speaking status to true", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)

  defer func() {
    err = voiceConnection.Speaking(false)

    if err != nil {
      logger.Error("playOggFile: Failed to set speaking status to false", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
      // TODO: Notify the user that there was an error
      return
    }

    logger.Debug("playOggFile: Set speaking status to false", "session", session, "interaction", interaction, "voiceConnection", voiceConnection)
  }()

  for {
    n, pkt, err := reader.ReadNextOpusPacket()
    if err == io.EOF {
      break
    }

    logger.Debug("playOggFile: Read next opus packet", "numBytesRead", n, "packet", pkt)

    if err != nil {
      // TODO: Notify the user that there was an error
      logger.Error("playOggFile: Error while reading most recent opus packet", "error", err)
      return
    }

    voiceConnection.OpusSend <- pkt
  }
}
