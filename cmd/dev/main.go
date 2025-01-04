package main

import (
	"context"
	"log"

	"accidentallycoded.com/fredboard/v3/audio/graph"
)

func main() {
	/*
		sink := graph.NewStdoutSinkNode()
		err := sink.Start(context.Background())
		if err != nil {
			log.Fatalln("sink.Start", err)
		}
	*/

	source := graph.NewFSFileSource()
	err := source.Open("sample.pcms16le", context.Background())
	if err != nil {
		log.Fatalln("source.Open", err)
	}

	sink := graph.NewFSFileSink()
	err = sink.Open("transcoded.pcms16le", context.Background())
	if err != nil {
		log.Fatalln("sink.Open", err)
	}

	source.AddChild(sink)
	err = sink.Start(context.Background())
	if err != nil {
		log.Fatalln("sink.Start", err)
	}

	err = sink.Wait()
	if err != nil {
		log.Fatalln("sink.Wait", err)
	}

	/*
		defer func() {
			err := source.Close()
			if err != nil {
				log.Fatalln("source.Close", err)
			}
		}()

		transcoder := graph.NewPCMS16LE_Opus_TranscoderNode()
		err = transcoder.Initialize()
		if err != nil {
			log.Fatalln("transcoder.Initialize", err)
		}

		transcoder.AddParent(source)
		transcoder.AddChild(sink)

		err = source.Wait()
		if err != context.Canceled {
			log.Fatalf("source.Wait", err)
		}
	*/
}
