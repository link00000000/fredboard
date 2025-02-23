package main

import (
	"bytes"
	"context"
	"os"

	"accidentallycoded.com/fredboard/v3/internal/audio/parallelgraph"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

func main() {
	logger := logging.NewLogger()
	logger.SetPanicOnError(true)
	logger.AddHandler(logging.NewPrettyHandler(os.Stderr, logging.LevelDebug))
	defer logger.Close()

	inBuf := bytes.NewBufferString("Hello, World!")
	var outBuf bytes.Buffer

	readerNode := parallelgraph.NewReaderNode(logger, inBuf)
	passthroughNodeOne := parallelgraph.NewPassthroughNode(logger)
	passthroughNodeTwo := parallelgraph.NewPassthroughNode(logger)
	writerNode := parallelgraph.NewWriterNode(logger, &outBuf)

	graph := parallelgraph.NewGraph(logger)
	graph.AddNode(readerNode)
	graph.AddNode(passthroughNodeOne)
	graph.AddNode(passthroughNodeTwo)
	graph.AddNode(writerNode)
	graph.CreateConnection(readerNode, passthroughNodeOne)
	graph.CreateConnection(passthroughNodeOne, passthroughNodeTwo)
	graph.CreateConnection(passthroughNodeTwo, writerNode)

	//ctx, cancel := context.WithCancel(context.Background())

	readerNode.OnEOF.AddDelegate(func(struct{}) {
		graph.FlushAndStop()
	})

	logger.Info("starting graph", "input", string(inBuf.Bytes()))
	graph.Start(context.Background())

	for err := range graph.Errors() {
		logger.Error("error from audio graph", "error", err)
	}

	logger.Info("done", "output", string(outBuf.Bytes()))
}
