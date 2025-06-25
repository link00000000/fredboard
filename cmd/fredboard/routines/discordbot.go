package routines

import (
	"accidentallycoded.com/fredboard/v3/internal/syncext"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

func discordBotRoutine(term <-chan bool) error {
	return nil
}

func NewDiscordBotRoutine(logger *logging.Logger, name string) syncext.Routine {
	return syncext.NewBasicRoutine(name, discordBotRoutine)
}
