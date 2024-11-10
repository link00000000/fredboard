package sources

import (
	"fmt"
	"testing"
)

func TestNewYouTubeStream(t *testing.T) {
  yt := NewYouTubeStream("https://www.youtube.com/watch?v=YWnJI5-fFJs", YOUTUBEQUALITY_WORST)

  dataChannel := make(chan []byte)
  errChannel := make(chan error);
  done := make(chan bool)

  go func() {
    /*
    for {
      select {
      case data := <-dataChannel:
        for i, b := range data {
          fmt.Printf("%02x ", b)

          if (i + 1) % (4 * 8) == 0 {
            fmt.Printf("\n")
          } else if (i + 1) % 8 == 0 {
            fmt.Printf("  ")
          }
        }
      case err := <-errChannel:
        t.Fatal("Got error from error channel", err)
      case <-done:
        fmt.Println("Done")
        break
      }
    }
    */

    for i, b := range <-dataChannel {
      fmt.Printf("%02x ", b)

      if (i + 1) % (4 * 8) == 0 {
        fmt.Printf("\n")
      } else if (i + 1) % 8 == 0 {
        fmt.Printf("  ")
      }
    }
  }()

  if err := yt.Start(dataChannel, errChannel); err != nil {
    t.Fatal("Failed to start YouTubeStream", err)
  } else {
    fmt.Println("Started")
  }

  if err := yt.Wait(); err != nil {
    t.Fatal("Failed to wait", err)
  } else {
    fmt.Println("Done waiting for yt")
    done <- true
  }
}
