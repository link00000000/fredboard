package main

import (
	"bytes"
	"context"
	"os"
	"sync"

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

	readerNode := parallelgraph.NewReaderNode(inBuf)
	writerNode := parallelgraph.NewWriterNode(&outBuf)

	graph := parallelgraph.NewGraph()
	graph.AddNode(readerNode)
	graph.AddNode(writerNode)
	graph.CreateConnection(readerNode, writerNode)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for err := range graph.Errors() {
			logger.Error("error from audio graph", "error", err)
		}
	}()

	logger.Debug("starting graph", "input", string(inBuf.Bytes()))
	graph.Start(context.TODO())
	wg.Wait()

	logger.Debug("done", "output", string(outBuf.Bytes()))
}
