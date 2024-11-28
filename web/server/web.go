package server

import (
	"net/http"

	"accidentallycoded.com/fredboard/v3/telemetry/logging"
)

type Web struct {
	Logger *logging.Logger
	Mux    *http.ServeMux
}

func NewWebServer(logger *logging.Logger) *Web {
	web := &Web{
		Logger: logger,
		Mux:    http.NewServeMux(),
	}

	return web
}

func (web *Web) NewLogger() *logging.Logger {
	return web.Logger.NewChildLogger()
}

// TODO
/*
func (web *Web) handleEventLogs(res http.ResponseWriter, req *http.Request) {
	logger, err := web.logger.NewChildLogger()
	if err != nil {
		logger.Fatal("failed to create child logger")
	}

	defer logger.Close()
	logger.SetData("request", &req)
	logger.SetData("response", &res)

	logger.Debug("received request")
	defer logger.Debug("closed request")

	id := web.logBroadcaster.AddResponse(res)
	defer web.logBroadcaster.RemoveResponse(id)

	// Leave the connection open until the client closes it
	// so they can receive notifications via SSE
	<-req.Context().Done()
}
*/
