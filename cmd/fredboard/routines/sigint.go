package routines

import (
	"os"
	"os/signal"

	"accidentallycoded.com/fredboard/v3/internal/syncext"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

func sigIntRoutine(logger *logging.Logger, term <-chan bool) error {
	defer logger.Info("stopping FredBoard")

	logger.Info("press ^c to exit")

	intSig := make(chan os.Signal, 1)
	signal.Notify(intSig, os.Interrupt)

	select {
	case <-intSig:
		logger.Debug("received interrupt signal")
		logger.Debug("SigIntRoutine requesting term of all routines")
		return syncext.ErrRequestTermAllRoutines
	case <-term:
		logger.Debug("SigIntRoutine received term request")
		return nil
	}
}

func NewSigIntRoutine(logger *logging.Logger, name string) syncext.Routine {
	return syncext.NewBasicRoutine(name, func(term <-chan bool) error { return sigIntRoutine(logger, term) })
}
