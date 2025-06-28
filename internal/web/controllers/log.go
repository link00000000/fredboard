package controllers

import (
	"html/template"
	"net/http"

	"github.com/link00000000/telemetry/logging"
	"github.com/link00000000/fredboard/v3/internal/web/content"
	"github.com/link00000000/fredboard/v3/internal/web/server"
)

type logsController struct {
	*Controller

	eventBroadcaster *server.SSEBroadcaster
}

func NewLogsController(srv *server.Web) *logsController {
	controller := &logsController{
		Controller:       newController(srv),
		eventBroadcaster: server.NewSSEBroadcaster(),
	}

	controller.mux.HandleFunc("/events", controller.handleEvents)
	controller.mux.HandleFunc("/", controller.handleIndex)

	// TODO: Handle levels
	srv.Logger.AddHandler(logging.NewJsonHandler(controller.eventBroadcaster, logging.LevelDebug))

	return controller
}

// Implements [http.Handler]
func (controller *logsController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	controller.mux.ServeHTTP(w, r)
}

func (controller *logsController) handleIndex(w http.ResponseWriter, r *http.Request) {
	logger := controller.newLoggerForRequest(w, r)
	defer logger.Close()

	templ, err := template.ParseFS(content.ContentFS, "templates/logs/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("failed to read template", "error", err)
		return
	}

	err = templ.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("failed to execute template", "error", err)
		return
	}
}

func (controller *logsController) handleEvents(w http.ResponseWriter, r *http.Request) {
	logger := controller.newLoggerForRequest(w, r)

	defer logger.Close()
	logger.SetData("request", &r)
	logger.SetData("response", &w)

	conn := server.NewSSEConnection(w)

	err := conn.EstablishConnection()
	if err == server.ErrStreamingUnsupported {
		http.Error(w, "streaming unsupported!", http.StatusInternalServerError)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	id := controller.eventBroadcaster.AddResponse(conn)
	defer controller.eventBroadcaster.RemoveResponse(id)

	// Leave the connection open until the client closes it
	// so they can receive notifications via SSE
	<-r.Context().Done()
}
