package controllers

import (
	"html/template"
	"net/http"

	"accidentallycoded.com/fredboard/v3/web/content"
	"accidentallycoded.com/fredboard/v3/web/server"
)

type logsController struct {
  *Controller
}

func NewLogsController(server *server.Web) *logsController {
  controller := &logsController{newController(server)}

  controller.mux.HandleFunc("/", controller.handleIndex)

  return controller
}

// Implements [http.Handler]
func (controller *logsController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  controller.mux.ServeHTTP(w, r)
}

func (controller *logsController) handleIndex(w http.ResponseWriter, r *http.Request) {
  logger := controller.newLoggerWithRequest(w, r)
	defer logger.Close()

	logger.Debug("received request")
	defer logger.Debug("closed request")

	templ, err := template.ParseFS(content.ContentFS, "templates/logs/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.ErrorWithErr("failed to read template", err)
    return
	}

	err = templ.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.ErrorWithErr("failed to execute template", err)
    return
	}
}
