package main

import (
	"context"
	"time"

	"accidentallycoded.com/fredboard/v3/audio/graph"
)

func main() {
	sink := graph.NewStdoutSinkNode()
	sinkCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	sink.Start(sinkCtx)

	source := graph.NewZeroSourceNode()

	passthrough := graph.NewPassthroughNode()
	passthrough.AddParent(source)
	passthrough.AddChild(sink)

	<-sinkCtx.Done()
}
