package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"accidentallycoded.com/fredboard/v3/internal/optional"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"accidentallycoded.com/fredboard/v3/internal/ytdlp"
)

func main() {
	logger := logging.NewLogger()
	logger.SetPanicOnError(true)
	logger.AddHandler(logging.NewPrettyHandler(os.Stdout, logging.LevelDebug))
	defer logger.Close()

	// metadata
	func() {
		metadataLogger := logger.NewChildLogger()
		defer metadataLogger.Close()

		metadata, err := ytdlp.GetMetadata(metadataLogger, "https://www.youtube.com/watch?v=jNQXAC9IVRw", &ytdlp.Config{
			ExePath:     optional.Make("/nix/store/8bk9vw8bk10x3g0r60mp1yxinfbwx4gd-yt-dlp-2025.1.26/bin/yt-dlp"),
			CookiesPath: optional.Make("/home/logan/Source/link00000000/fredboard/youtube.com_cookies.txt"),
		}, context.TODO())

		if err != nil {
			logger.Panic("failed to get metadata", "error", err)
		}

		logger.Info("done", "metadata", metadata)
		f, err := os.Create("./metadata.json")
		if err != nil {
			logger.Panic("failed to create metadata output file", "error", err)
		}
		defer f.Close()

		b, err := json.MarshalIndent(metadata, "", "    ")
		if err != nil {
			logger.Panic("failed to marshal metadata", "error", err)
		}

		_, err = f.Write(b)
		if err != nil {
			logger.Panic("failed to write metdata to file", "error", err)
		}
	}()

	// video
	func() {
		videoLogger := logger.NewChildLogger()
		defer videoLogger.Close()

		_, _ = context.WithTimeout(context.Background(), time.Second*2)

		videoReader, err := ytdlp.Open(videoLogger, "https://www.youtube.com/watch?v=jNQXAC9IVRw", &ytdlp.Config{
			ExePath:     optional.Make("/nix/store/8bk9vw8bk10x3g0r60mp1yxinfbwx4gd-yt-dlp-2025.1.26/bin/yt-dlp"),
			CookiesPath: optional.Make("/home/logan/Source/link00000000/fredboard/youtube.com_cookies.txt"),
		}, context.TODO())

		if err != nil {
			logger.Panic("failed to get video data", "error", err)
		}

		logger.Info("started video stream")

		f, err := os.Create("./video.mkv")
		if err != nil {
			logger.Panic("failed to create video output file", "error", err)
		}
		defer f.Close()

		buf := make([]byte, 0xffff)
		for {
			nr, err := videoReader.Read(buf)
			buf = buf[:nr]

			if err == io.EOF {
				break
			}

			if err != nil {
				logger.Panic("failed to read video data", "error", err)
			}

			nw, err := f.Write(buf)
			if err != nil {
				logger.Panic("failed to write video data to output file", "error", err)
			}

			if nr != nw {
				panic("nr != nw")
			}
		}

		logger.Info("done writing video data")
	}()
}
