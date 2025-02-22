package extensions

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"accidentallycoded.com/fredboard/v3/internal/audio/graph"
	"accidentallycoded.com/fredboard/v3/internal/exec/ffmpeg"
	"accidentallycoded.com/fredboard/v3/internal/exec/ytdlp"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

var (
	_ graph.AudioGraphNode = (*YouTubeSourceNode)(nil)
)

type YouTubeSourceNode struct {
	buf    bytes.Buffer
	logger *logging.Logger
	m      sync.Mutex
}

func NewYouTubeSourceNode(logger *logging.Logger) *YouTubeSourceNode {
	return &YouTubeSourceNode{
		logger: logger,
	}
}

func (node *YouTubeSourceNode) OpenVideo(ytdlpConfig *ytdlp.Config, ffmpegConfig *ffmpeg.Config, url string, quality ytdlp.YtdlpAudioQuality) (cDone chan struct{}, cErr chan error, err error) {
	cDone = make(chan struct{})
	cErr = make(chan error)

	ctx := context.TODO()

	ytdlpCmd, err := ytdlp.NewVideoCmd(ctx, ytdlpConfig, url, ytdlp.YtdlpAudioQuality_BestAudio)
	if err != nil {
		node.logger.Warn("error while executing ytdlp.NewVideoCmd", "error", err, "url", url)
		return nil, nil, fmt.Errorf("error while executing ytdlp.NewVideoCmd: %w", err)
	}

	ytdlpStdout, err := ytdlpCmd.StdoutPipe()
	if err != nil {
		node.logger.Warn("error while creating stdout pipe for subprocess yt-dlp", "error", err)
		return nil, nil, fmt.Errorf("error while creating stdout pipe for subprocess yt-dlp: %w", err)
	}

	ytdlpStderr, err := ytdlpCmd.StderrPipe()
	if err != nil {
		node.logger.Warn("error while creating stderr pipe for subprocess yt-dlp", "error", err)
		return nil, nil, fmt.Errorf("error while creating stderr pipe for subprocess yt-dlp: %w", err)
	}

	ffmpegCmd, err := ffmpeg.NewEncodeCmd(ctx, ffmpegConfig, ffmpeg.Format_Ogg, 48000, 2)
	if err != nil {
		node.logger.Warn("error while executing ffmpeg.NewEncodeCmd", "error", err)
		return nil, nil, fmt.Errorf("error while executing ffmpeg.NewEncodeCmd: %w", err)
	}

	ffmpegStdout, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		node.logger.Warn("error while creating stdout pipe for subprocess ffmpeg", "error", err)
		return nil, nil, fmt.Errorf("error while creating stdout pipe for subprocess ffmpeg: %w", err)
	}

	ffmpegStderr, err := ffmpegCmd.StderrPipe()
	if err != nil {
		node.logger.Warn("error while creating stderr pipe for subprocess ffmpeg", "error", err)
		return nil, nil, fmt.Errorf("error while creating stderr pipe for subprocess ffmpeg: %w", err)
	}

	ffmpegCmd.Stdin = ytdlpStdout

	if ytdlpCmd.Start(); err != nil {
		node.logger.Warn("error while starting yt-dlp subprocess", "error", err)
		return nil, nil, fmt.Errorf("error while starting yt-dlp subprocess: %w", err)
	} else {
		node.logger.Debug("started yt-dlp subprocess", "exe", ytdlpCmd.Path, "args", ytdlpCmd.Args)
	}

	if err = ffmpegCmd.Start(); err != nil {
		node.logger.Warn("error while starting ffmpeg subprocess", "error", err)
		return nil, nil, fmt.Errorf("error while starting ffmpeg subprocess: %w", err)
	} else {
		node.logger.Debug("started ffmpeg subprocess", "exe", ffmpegCmd.Path, "args", ffmpegCmd.Args)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		node.logger.Debug("starting goroutine to copy ffmpeg stdout to node buffer")

		defer func() {
			wg.Done()
			node.logger.Debug("done copying ffmeg stdout to node buffer")
		}()

		internalBuf := make([]byte, 0x8000)
		for {
			n, err := ffmpegStdout.Read(internalBuf)
			internalBuf = internalBuf[:n]
			if err == io.EOF {
				break
			}

			if err != nil {
				node.logger.Warn("error while copying ffmpeg stdout to internal buffer", "error", err)
				cErr <- fmt.Errorf("error while copying ffmpeg stdout to internal buffer: %w", err)
				return
			}

			node.m.Lock()
			node.buf.Write(internalBuf)
			node.m.Unlock()

			if err != nil {
				node.logger.Warn("internal buffer to node buffer", "error", err)
				cErr <- fmt.Errorf("internal buffer to node buffer: %w", err)
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		node.logger.Debug("starting goroutine to copy yt-dlp stdout to ffmpeg stdin")

		defer func() {
			wg.Done()
			node.logger.Debug("done copying yt-dlp stdout to ffmpeg stdin")
		}()

		if err := node.logger.LogReader(ytdlpStderr, logging.LevelDebug, "[ytdlp stderr] %s"); err != nil {
			node.logger.Warn("error while copying yt-dlp stdout to ffmpeg stdin", "error", err)
			cErr <- fmt.Errorf("error while copying yt-dlp stdout to ffmpeg stdin: %w", err)
		}
	}()

	wg.Add(1)
	go func() {
		node.logger.Debug("starting goroutine to read stderr from ffmpeg")

		defer func() {
			wg.Done()
			node.logger.Debug("done reading stderr from ffmpeg to complete")
		}()

		if err := node.logger.LogReader(ffmpegStderr, logging.LevelDebug, "[ffmpeg stderr] %s"); err != nil {
			node.logger.Warn("error while reading stderr from ffmpeg")
			cErr <- fmt.Errorf("error while reading stder from ffmpeg: %w", err)
		}
	}()

	wg.Add(1)
	go func() {
		node.logger.Debug("starting goroutine to wait for yt-dlp subprocess to complete")

		defer func() {
			wg.Done()
			node.logger.Debug("done waiting for yt-dlp subprocess to complete")
		}()

		err := ytdlpCmd.Wait()
		node.logger.Debug("yt-dlp subprocess exited with code", "exitCode", ytdlpCmd.ProcessState.ExitCode())

		if err != nil {
			node.logger.Warn("error while waiting for yt-dlp subprocess", "error", err)
			cErr <- fmt.Errorf("error while waiting for yt-dlp subprocess: %w", err)

			if err := ffmpegCmd.Process.Kill(); err != nil {
				node.logger.Warn("error while killing ffmpeg process", "error", err)
				cErr <- fmt.Errorf("error while killing ffmpeg process: %w", err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		node.logger.Debug("starting goroutine to wait for ffmpeg subprocess to complete")

		defer func() {
			wg.Done()
			node.logger.Debug("done waiting for ffmpeg subprocess")
		}()

		err := ffmpegCmd.Wait()
		node.logger.Debug("ffmpg subprocess exited with code", "exitCode", ffmpegCmd.ProcessState.ExitCode())

		if err != nil {
			node.logger.Warn("error while waiting for ffmpeg subprocess", "error", err)
			cErr <- fmt.Errorf("error while waiting for ffmpeg subprocess: %w", err)
		}
	}()

	go func() {
		node.logger.Debug("starting goroutine to wait for all items in wait group")

		defer func() {
			close(cErr)
			close(cDone)
			node.logger.Debug("done waiting for all items in wait group")
		}()

		node.logger.Debug("waiting for all items in wait group")
		wg.Wait()
	}()

	return cDone, cErr, nil
}

// Implements [nodes.AudioGraphNode]
func (node *YouTubeSourceNode) Tick(ins []io.Reader, outs []io.Writer) error {
	if err := graph.AssertNodeIOBounds(ins, graph.NodeIOType_In, 0, 0); err != nil {
		return fmt.Errorf("DiscordSinkNode.Tick error: %w", err)
	}

	if err := graph.AssertNodeIOBounds(outs, graph.NodeIOType_Out, 1, 1); err != nil {
		return fmt.Errorf("DiscordSinkNode.Tick error: %w", err)
	}

	node.m.Lock()
	_, err := node.buf.WriteTo(outs[0])
	node.m.Unlock()

	if err != nil {
		return fmt.Errorf("YouTubeSourceNode.Tick() failed to write to next node in audio graph: %w", err)
	}

	return nil
}
