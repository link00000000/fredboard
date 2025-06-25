//go:build !debug

package main

import (
	"accidentallycoded.com/fredboard/v3/cmd/fredboard/routines"
	"accidentallycoded.com/fredboard/v3/internal/syncext"
)

func main() {
	logger := Init()

	/*
		if err := libgui.Load("bin/libgui.so"); err != nil {
			panic(err)
		}

		for {
		}
	*/

	routineManager := syncext.NewRoutineManager()

	routineManager.StartRoutine(routines.NewUIRoutine(logger, "ui"))
	routineManager.StartRoutine(routines.NewDiscordBotRoutine(logger, "discord bot"))
	routineManager.StartRoutine(routines.NewSigIntRoutine(logger, "sigint"))

	routineManager.WaitForAllRoutines()
}
