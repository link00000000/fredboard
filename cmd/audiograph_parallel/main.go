package main

import (
	"bytes"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

	"accidentallycoded.com/fredboard/v3/internal/audio/parallelgraph"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

func main() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	runtime.SetMutexProfileFraction(16)
	runtime.SetBlockProfileRate(16)

	logger := logging.NewLogger()
	logger.SetPanicOnError(true)
	logger.AddHandler(logging.NewPrettyHandler(os.Stderr, logging.LevelDebug))
	defer logger.Close()

	inBuf := bytes.NewBufferString("")
	var outBuf bytes.Buffer

	readerNode := parallelgraph.NewReaderNode(logger, inBuf)
	writerNode := parallelgraph.NewWriterNode(logger, &outBuf)

	graph := parallelgraph.NewGraph(logger)
	graph.AddNode(readerNode)
	graph.AddNode(writerNode)
	graph.CreateConnection(readerNode, writerNode)

	readerNode.OnEOF.AddDelegate(func(struct{}) {
		go func() {
			logger.Info("eof reached")
			logger.Info("done", "output", string(outBuf.Bytes()))
			graph.Stop(parallelgraph.FlushPolicy_Flush)
		}()
	})

	logger.Info("starting graph", "input", string(inBuf.Bytes()))
	graph.Start()

	for err := range graph.Errors() {
		logger.Error("error from audio graph", "error", err)
	}
}
