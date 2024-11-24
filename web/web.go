package web

import (
	"fmt"
	"net/http"
	"time"

	"accidentallycoded.com/fredboard/v3/telemetry"
)

func Start(ltx *telemetry.Context) {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(res, "<h1>Hello, world!</h1>")
	})

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Flush the headers to establish the connection
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

    select {
      
    }

		// Send events in a loop
		for {
			// Create a message
			message := fmt.Sprintf("data: Current time is: %s\n\n", time.Now().Format(time.RFC3339))

			// Write the message to the response
			fmt.Fprint(w, message)

			// Flush the data to the client
			flusher.Flush()

			// Wait for a second before sending the next event
			time.Sleep(1 * time.Second)
		}
	})

	http.ListenAndServe(":8080", nil)
}
