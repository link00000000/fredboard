package server

import (
	"net/http"

	"github.com/link00000000/fredboard/v3/internal/telemetry/logging"
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

// Implements [http.Handler]
func (web *Web) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := web.Logger.NewChildLogger()

	logger.SetData("request.method", r.Method)
	logger.SetData("request.url", r.URL)

	logger.Debug("received request")
	defer logger.Debug("closed request")

	web.Mux.ServeHTTP(w, r)
}
