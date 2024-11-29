package web

import (
	"net/http"

	"accidentallycoded.com/fredboard/v3/telemetry/logging"
	"accidentallycoded.com/fredboard/v3/web/controllers"
	"accidentallycoded.com/fredboard/v3/web/server"
)

func Start(address string, logger *logging.Logger) {
	server := server.NewWebServer(logger)

	logsController := controllers.NewLogsController(server)
	defer logsController.Close()
	server.Mux.Handle("/logs/", http.StripPrefix("/logs", logsController))

	staticController := controllers.NewStaticController(server)
	defer staticController.Close()
	server.Mux.Handle("/static/", http.StripPrefix("/static", staticController))

	logger.SetData("web", server)

	logger.Info("listening for requests")
	http.ListenAndServe(address, server)
}
