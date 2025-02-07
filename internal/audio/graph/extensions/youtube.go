package extensions

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"accidentallycoded.com/fredboard/v3/internal/audio/graph"
	"accidentallycoded.com/fredboard/v3/internal/events"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var (
	_ graph.AudioGraphNode = (*YouTubeSourceNode)(nil)
)

type YouTubeStreamQuality string

const (
	YOUTUBESTREAMQUALITY_WORST YouTubeStreamQuality = "worstaudio"
	YOUTUBESTREAMQUALITY_BEST                       = "bestaudio"
)

type YouTubeSourceNode struct {
	OnDoneEvent *events.EventEmitter[error]

	ytdlpProc  *process
	ffmpegProc *process

	logger *logging.Logger
}

type process struct {
	Cmd    *exec.Cmd
	Stdout io.ReadCloser
}

func NewYouTubeSourceNode(logger *logging.Logger) *YouTubeSourceNode {
	return &YouTubeSourceNode{
		OnDoneEvent: events.NewEventEmitter[error](),

		ytdlpProc:  nil,
		ffmpegProc: nil,

		logger: logger,
	}
}

func (node *YouTubeSourceNode) OpenVideo(url string, quality YouTubeStreamQuality) error {
	ytdlpProc, err := node.newYtdlpProc(url, quality)
	if err != nil {
		return fmt.Errorf("YouTubeSourceNode.OpenVideo() failed to create yt-dlp process: %w", err)
	}
	node.ytdlpProc = ytdlpProc

	ffmpegProc, err := node.newFfmpegProc()
	if err != nil {
		return fmt.Errorf("YouTubeSourceNode.OpenVideo() failed to create ffmpeg process: %w", err)
	}
	node.ffmpegProc = ffmpegProc

	ffmpegProc.Cmd.Stdin = ytdlpProc.Stdout

	err = ytdlpProc.Cmd.Start()
	if err != nil {
		return fmt.Errorf("YouTubeSourceNode.OpenVideo() failed to start yt-dlp process: %w", err)
	}

	err = ffmpegProc.Cmd.Start()
	if err != nil {
		return fmt.Errorf("YouTubeSourceNode.OpenVideo() failed to start ffmpeg process: %w", err)
	}

	go func() {
		// TODO
		errs := make([]error, 0, 2)

		//err := ytdlpProc.Cmd.Wait()
		if err != nil {
			errs = append(errs, err)
		}

		//err = ffmpegProc.Cmd.Wait()
		if err != nil {
			errs = append(errs, err)
		}

		switch len(errs) {
		case 0:
			err = nil
		case 1:
			err = errs[0]
		default:
			err = errors.Join(errs...)
		}

		//node.OnDoneEvent.Broadcast(err)
	}()

	return nil
}

// Implements [nodes.AudioGraphNode]
func (node *YouTubeSourceNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := graph.AssertNodeIOBounds(ins, graph.NodeIOType_In, 0, 0); err != nil {
		return fmt.Errorf("DiscordSinkNode.Tick error: %w", err)
	}

	if err := graph.AssertNodeIOBounds(outs, graph.NodeIOType_Out, 1, 1); err != nil {
		return fmt.Errorf("DiscordSinkNode.Tick error: %w", err)
	}

	buf := [0xffff]byte{}
	n, err := node.ffmpegProc.Stdout.Read(buf[:])
	if err != nil {
		return fmt.Errorf("YouTubeSourceNode.Tick() failed to read from ffmpeg process: %w", err)
	}

	_, err = outs[0].Write(buf[:n])
	if err != nil {
		return fmt.Errorf("YouTubeSourceNode.Tick() failed to write to next node in audio graph: %w", err)
	}

	return nil
}

func (node *YouTubeSourceNode) Stop() error {
	// TODO
	return nil
	errs := make([]error, 0, 2)

	err := node.ytdlpProc.Cmd.Process.Kill()
	if err != nil {
		errs = append(errs, err)
	}

	err = node.ffmpegProc.Cmd.Process.Kill()
	if err != nil {
		errs = append(errs, err)
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return errors.Join(errs...)
	}
}

func (node *YouTubeSourceNode) newYtdlpProc(url string, quality YouTubeStreamQuality) (*process, error) {
	executablePath, err := exec.LookPath("yt-dlp")
	if err != nil {
		return nil, fmt.Errorf("YouTubeSourceNode.newYtdlpProc() failed to find path to yt-dlp executable: %w", err)
	}

	proc := exec.Command(executablePath,
		"--abort-on-error",
		//"--quiet",
		"--format", fmt.Sprintf("%s[acodec=opus]", quality),
		"--output", "-",
		url)

	proc.Stderr = os.Stdout
	//stderr, err := proc.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("YouTubeSourceNode.newYtdlpProc() failed to create stderr pipe: %w", err)
	}

	// log stderr
	go func() {
		buf := make([]byte, 0, 0xff)

		for {
			buf = buf[:0]
			//_, err := stderr.Read(buf)

			if errors.Is(err, io.EOF) || errors.Is(err, os.ErrClosed) {
				return
			}

			if err != nil {
				node.logger.ErrorWithErr("YouTubeSourceNode.newYtdlpProc() failed to read from stderr pipe", err)
				return
			}

			if len(buf) > 0 {
				node.logger.Warn(fmt.Sprintf("ytdlp stderr: %s", string(buf)))
			}
		}
	}()

	stdout, err := proc.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("YouTubeSourceNode.newYtdlpProc() failed to create stdout pipe: %w", err)
	}

	return &process{proc, stdout}, nil
}

func (node *YouTubeSourceNode) newFfmpegProc() (*process, error) {
	executablePath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("YouTubeSourceNode.newFfmpegProc() failed to find path to ffmpeg executable: %w", err)
	}

	proc := exec.Command(executablePath,
		"-hide_banner",
		//"-loglevel", "error",
		"-i", "pipe:0",
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"pipe:1")

	proc.Stderr = os.Stdout
	//stderr, err := proc.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("YouTubeSourceNode.newFfmpegProc() failed to create stderr pipe: %w", err)
	}

	go func() {
		buf := make([]byte, 0, 0xffff)

		for {
			buf = buf[:0]
			//_, err := stderr.Read(buf)

			if errors.Is(err, io.EOF) || errors.Is(err, os.ErrClosed) {
				return
			}

			if err != nil {
				node.logger.ErrorWithErr("YouTubeSourceNode.newFfmpegProc() failed to read from stderr pipe", err)
				return
			}

			if len(buf) > 0 {
				node.logger.Warn(fmt.Sprintf("ffmpeg stderr: %s", string(buf)))
			}
		}
	}()

	stdout, err := proc.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("YouTubeSourceNode.newFfmpegProc() failed to create stdout pipe: %w", err)
	}

	return &process{proc, stdout}, nil
}
