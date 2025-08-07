package audiosession

import (
	"context"
	"fmt"

	"accidentallycoded.com/fredboard/v3/internal/audio"
	"accidentallycoded.com/fredboard/v3/internal/config"
	"accidentallycoded.com/fredboard/v3/internal/exec/ffmpeg"
	"accidentallycoded.com/fredboard/v3/internal/exec/ytdlp"
	"accidentallycoded.com/fredboard/v3/internal/telemetry"
)

type YtdlpInput struct {
	*BaseInput
}

func (i *YtdlpInput) Pause() {
	i.BaseInput.Pause()

	// TODO
	panic("unimplemeted")
}

func (i *YtdlpInput) Resume() {
	i.BaseInput.Resume()

	// TODO
	panic("unimplemeted")
}

// add a ytdlp input that will automatically be stopped when EOF is reached
func (s *Session) AddYtdlpInput(ctx context.Context, url string, quality ytdlp.YtdlpAudioQuality) (Input, error) {
	videoReader, err, videoReaderExitChan := ytdlp.NewVideoReader(ytdlp.Config{ExePath: config.Get().Ytdlp.ExePath, CookiesPath: config.Get().Ytdlp.CookiesFile}, url, quality)

	if err != nil {
		return nil, fmt.Errorf("failed to create video reader: %w", err)
	}

	transcoder, err, transcoderExitChan := ffmpeg.NewTranscoder(
		ctx,
		ffmpeg.Config{ExePath: config.Get().Ffmpeg.ExePath},
		videoReader,
		ffmpeg.Format_PCMSigned16BitLittleEndian,
		config.Get().Audio.SampleRateHz,
		config.Get().Audio.NumChannels,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create transcoder: %w", err)
	}

	// TODO: Put 0x8000 in config
	videoReaderNode := audio.NewReaderNode(transcoder, 0x8000)

	input := &YtdlpInput{BaseInput: NewBaseInput(s, videoReaderNode)}
	s.AddInput(input)

	go func() {
		err := <-videoReaderExitChan
		if err != nil {
			telemetry.Logger.ErrorContext(ctx, "ytdlp videoReader exited with exit error", "err", err)
		} else {
			telemetry.Logger.InfoContext(ctx, "ytdlp videoReader exited successfully")
		}

		err = <-transcoderExitChan
		if err != nil {
			telemetry.Logger.ErrorContext(ctx, "ytdlp transcoder exited with exit error", "err", err)
		} else {
			telemetry.Logger.InfoContext(ctx, "ytdlp transcoder exited successfully")
		}

		input.Stop()
	}()

	return input, nil
}
