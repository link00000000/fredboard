package controllers

import (
	"net/http"

	"accidentallycoded.com/fredboard/v3/internal/web/content"
	"accidentallycoded.com/fredboard/v3/internal/web/server"
)

type staticController struct {
	*Controller
}

func NewStaticController(srv *server.Web) *staticController {
	controller := &staticController{newController(srv)}

	controller.mux.HandleFunc("/", controller.handleIndex)

	return controller
}

// Implements [http.Handler]
func (controller *staticController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	controller.mux.ServeHTTP(w, r)
}

// Implements [io.Closer]
func (controller *staticController) Close() error {
	return controller.logger.Close()
}

func (controller *staticController) handleIndex(w http.ResponseWriter, r *http.Request) {
	logger := controller.newLoggerForRequest(w, r)

	path := "static" + r.URL.Path

	logger.SetData("path", path)
	logger.SetData("fs", content.ContentFS)

	logger.Debug("Serving static asset")
	http.ServeFileFS(w, r, content.ContentFS, path)
}
