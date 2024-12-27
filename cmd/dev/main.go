package main

import (
	"context"

	"accidentallycoded.com/fredboard/v3/audio/graph"
)

func main() {
	sink := graph.NewStdoutSinkNode()
	sink.Start(context.Background())

	source := graph.NewFSFileSource()
	source.Open("./sample.ogg", context.Background())
	defer source.Close()

	passthrough := graph.NewPassthroughNode()
	passthrough.AddParent(source)
	passthrough.AddChild(sink)

	source.Wait()
}
