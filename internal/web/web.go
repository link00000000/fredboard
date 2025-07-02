package web

import (
	"context"
	"net/http"

	"github.com/link00000000/go-telemetry/logging"
	"github.com/link00000000/fredboard/v3/internal/web/controllers"
	"github.com/link00000000/fredboard/v3/internal/web/server"
)

func Run(ctx context.Context, address string, logger *logging.Logger) {
	server := server.NewWebServer(logger)

	logsController := controllers.NewLogsController(server)
	defer logsController.Close()
	server.Mux.Handle("/logs/", http.StripPrefix("/logs", logsController))

	staticController := controllers.NewStaticController(server)
	defer staticController.Close()
	server.Mux.Handle("/static/", http.StripPrefix("/static", staticController))

	logger.SetData("web", server)

	httpServer := &http.Server{Addr: address, Handler: server}

	go func() {
		logger.Info("listening for requests")
		err := httpServer.ListenAndServe()

		logger.Info("http server closed")

		if err != http.ErrServerClosed {
			logger.Error("http server error", "error", err)
		}
	}()

	<-ctx.Done()
	logger.Info("stopping web server")

	err := httpServer.Shutdown(ctx)
	logger.Info("http server shutdown")

	if err != nil {
		logger.Error("error while stopping http server", "error", err)
	}
}
