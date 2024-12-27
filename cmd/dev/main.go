package main

import (
	"log"
	"time"

	"accidentallycoded.com/fredboard/v3/audio/graph"
)

func main() {
	sink := graph.NewStdoutSinkNode()

	passthrough := graph.NewPassthroughNode()
	passthrough.AddChild(sink)

	source := graph.NewZeroSourceNode()
	source.AddChild(passthrough)

	err := sink.Start()
	if err != nil {
		log.Fatalln("Failed to start audio graph", err)
	}

	go func() {
		time.Sleep(5 * time.Minute)

		err := sink.Stop()
		if err != nil {
			log.Fatalln("Failed to stop audio graph", err)
		}
	}()

	sink.Wait()
}
