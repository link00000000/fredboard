package controllers

import (
	"net/http"

	"github.com/link00000000/fredboard/v3/internal/telemetry/logging"
	"github.com/link00000000/fredboard/v3/internal/web/server"
)

type Controller struct {
	server *server.Web
	mux    *http.ServeMux
	logger *logging.Logger
}

func newController(server *server.Web) *Controller {
	controller := &Controller{
		server: server,
		mux:    http.NewServeMux(),
		logger: server.Logger.NewChildLogger(),
	}

	return controller
}

// Implements [io.Closer]
func (controller *Controller) Close() error {
	return controller.logger.Close()
}

func (controller *Controller) newLogger() *logging.Logger {
	return controller.logger.NewChildLogger()
}

func (controller *Controller) newLoggerForRequest(w http.ResponseWriter, r *http.Request) *logging.Logger {
	logger := controller.newLogger()

	logger.SetData("request", &r)
	logger.SetData("response", &w)

	return logger
}
