//go:build debug

package main

import (
	"sync"

	"accidentallycoded.com/fredboard/v3/cmd/fredboard/routines"
	"accidentallycoded.com/fredboard/v3/internal/syncext"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

func UIHotReloadRoutine(logger *logging.Logger, routineManager *syncext.RoutineManager, term <-chan bool) error {
	m := syncext.NewRoutineManager()
	m.StartRoutine(routines.NewUIRoutine(logger, "ui hot reload - ui"))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		<-term
		m.TerminateAllRoutines(false, nil)
	}()

	m.WaitForAllRoutines()
	wg.Wait()

	return nil
}

func main() {
	logger := Init()

	routineManager := syncext.NewRoutineManager()

	routineManager.StartRoutine(syncext.NewBasicRoutine("ui hot reload", func(term <-chan bool) error { return UIHotReloadRoutine(logger, routineManager, term) }))
	//routineManager.StartRoutine(routines.NewUIRoutine(logger, "ui"))
	routineManager.StartRoutine(routines.NewDiscordBotRoutine(logger, "discord bot"))
	routineManager.StartRoutine(routines.NewSigIntRoutine(logger, "sigint"))

	routineManager.WaitForAllRoutines()
}
