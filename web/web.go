package web

import (
	"fmt"
	"net/http"
	"time"

	"accidentallycoded.com/fredboard/v3/telemetry/logging"
)

type Web struct {
	address string
	logger  *logging.Logger
}

func NewWeb(logger *logging.Logger) Web {
	return Web{address: ":8080", logger: logger}
}

func (web *Web) handleRoot(res http.ResponseWriter, req *http.Request) {
	logger, err := web.logger.NewChildLogger()
	if err != nil {
		logger.Fatal("failed to create child logger for web.handleRoot")
	}

	defer logger.Close()
	logger.SetData("request", &req)
	logger.SetData("response", &res)

	logger.Debug("Received request")
	defer logger.Debug("Closed request")

	fmt.Fprintf(res, "<h1>Hello, world!</h1>")
}

func (web *Web) handleLogs(res http.ResponseWriter, req *http.Request) {
	logger, err := web.logger.NewChildLogger()
	if err != nil {
		logger.Fatal("failed to create child logger for web.handleRoot")
	}

	defer logger.Close()
	logger.SetData("request", &req)
	logger.SetData("response", &res)

	logger.Debug("Received request")
	defer logger.Debug("Closed request")

	// Set headers for SSE
	res.Header().Set("Content-Type", "text/event-stream")
	res.Header().Set("Cache-Control", "no-cache")
	res.Header().Set("Connection", "keep-alive")

	// Flush the headers to establish the connection
	flusher, ok := res.(http.Flusher)
	if !ok {
		http.Error(res, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Send events in a loop
	for {
		// Create a message
		message := fmt.Sprintf("data: Current time is: %s\n\n", time.Now().Format(time.RFC3339))

		// Write the message to the response
		fmt.Fprint(res, message)

		// Flush the data to the client
		flusher.Flush()

		// Wait for a second before sending the next event
		time.Sleep(1 * time.Second)
	}
}

func (web *Web) Start() {
	web.logger.SetData("web", web)

	registerRoutesLogger, err := web.logger.NewChildLogger()
	if err != nil {
		web.logger.FatalWithErr("failed to create logger for web.registerRoutes", err)
	}

	defer registerRoutesLogger.Close()

	http.HandleFunc("/", web.handleRoot)
	registerRoutesLogger.SetData("route", "/")
	registerRoutesLogger.Debug("Registered route")

	http.HandleFunc("/logs", web.handleLogs)
	registerRoutesLogger.SetData("route", "/logs")
	registerRoutesLogger.Debug("Registered route")

	web.logger.Info("Listening for requests")
	http.ListenAndServe(web.address, nil)
}
