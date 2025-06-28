package audiosession

import (
	"fmt"

	"github.com/link00000000/fredboard/v3/pkg/gaps"
	"github.com/link00000000/fredboard/v3/internal/config"
	"github.com/link00000000/fredboard/v3/internal/exec/ffmpeg"
	"github.com/link00000000/fredboard/v3/internal/exec/ytdlp"
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
func (s *Session) AddYtdlpInput(url string, quality ytdlp.YtdlpAudioQuality) (Input, error) {
	videoReader, err, videoReaderExitChan := ytdlp.NewVideoReader(s.logger, ytdlp.Config{ExePath: config.Get().Ytdlp.ExePath, CookiesPath: config.Get().Ytdlp.CookiesFile}, url, quality)

	if err != nil {
		return nil, fmt.Errorf("failed to create video reader: %w", err)
	}

	transcoder, err, transcoderExitChan := ffmpeg.NewTranscoder(
		s.logger,
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
	videoReaderNode := audio.NewReaderNode(s.logger, transcoder, 0x8000)

	input := &YtdlpInput{BaseInput: NewBaseInput(s, videoReaderNode)}
	s.AddInput(input)

	go func() {
		err := <-videoReaderExitChan
		if err != nil {
			s.logger.Error("ytdlp videoReader exited with exit error", "err", err)
		} else {
			s.logger.Debug("ytdlp videoReader exited successfully")
		}

		err = <-transcoderExitChan
		if err != nil {
			s.logger.Error("ytdlp transcoder exited with exit error", "err", err)
		} else {
			s.logger.Debug("ytdlp transcoder exited successfully")
		}

		input.Stop()
	}()

	return input, nil
}
