package sources

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
)

type YouTubeQuality string

const (
	YOUTUBEQUALITY_WORST YouTubeQuality = "worstaudio"
	YOUTUBEQUALITY_BEST                 = "bestaudio"
)

type YouTubeStream struct {
	url     string
	quality YouTubeQuality

	dataChannel chan []byte
	errChannel  chan error

	ytdlpCmd  *exec.Cmd
	ffmpegCmd *exec.Cmd
}

func NewYouTubeStream(url string, quality YouTubeQuality) *YouTubeStream {
	return &YouTubeStream{url: url, quality: quality}
}

func (s *YouTubeStream) Start(dataChannel chan []byte, errChannel chan error) error {
	s.ytdlpCmd = exec.Command("yt-dlp",
		"--abort-on-error",
		"--quiet",
		"--no-warnings",
		"--format", fmt.Sprintf("%s[acodec=opus]", s.quality),
		"--output", "-",
		s.url)

	s.ffmpegCmd = exec.Command("ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-i", "pipe:0",
		"-c:a", "libopus",
		"-f", "opus",
		"pipe:1")

	if ytdlpStdout, err := s.ytdlpCmd.StdoutPipe(); err != nil {
		return err
	} else {
		s.ffmpegCmd.Stdin = ytdlpStdout
	}

	if ytdlpStderr, err := s.ytdlpCmd.StderrPipe(); err != nil {
		return err
	} else {
		go func() {
			for {
				buf := make([]byte, 0xff)
				if _, err := ytdlpStderr.Read(buf); err == io.EOF || err == io.ErrUnexpectedEOF {
					return
				} else if err != nil {
					errChannel <- err
					return
				} else {
					errChannel <- errors.New(fmt.Sprintf("ytdlp stderr: %s", string(buf)))
				}
			}
		}()
	}

	if _, err := s.ffmpegCmd.StdoutPipe(); err != nil {
		return err
	} else {
		go func() {
			/*
			   for {
			     reader := codecs.NewOggOpusReader(ffmpegStdout)
			     if _, pkt, err := reader.ReadNextPacket(); err == io.EOF || err == io.ErrUnexpectedEOF {
			       return
			     } else if err != nil {
			       errChannel <- err
			       return
			     } else {
			       for _, s := range pkt.Segments {
			         dataChannel <- s
			       }
			     }
			   }
			*/
		}()
	}

	if ffmpegStderr, err := s.ffmpegCmd.StderrPipe(); err != nil {
		return err
	} else {
		go func() {
			for {
				buf := make([]byte, 0xff)
				if _, err := ffmpegStderr.Read(buf); err == io.EOF || err == io.ErrUnexpectedEOF {
					return
				} else if err != nil {
					errChannel <- err
					return
				} else {
					errChannel <- errors.New(fmt.Sprintf("ffmpeg stderr: %s", string(buf)))
				}
			}
		}()
	}

	if err := s.ytdlpCmd.Start(); err != nil {
		return err
	}

	if err := s.ffmpegCmd.Start(); err != nil {
		return err
	}

	return nil
}

func (s *YouTubeStream) Stop() error {
	if err := s.ytdlpCmd.Process.Kill(); err != nil {
		return err
	}

	if err := s.ffmpegCmd.Process.Kill(); err != nil {
		return err
	}

	return nil
}

func (s *YouTubeStream) Wait() error {
	if err := s.ytdlpCmd.Wait(); err != nil {
		return err
	}

	if err := s.ffmpegCmd.Wait(); err != nil {
		return err
	}

	return nil
}
