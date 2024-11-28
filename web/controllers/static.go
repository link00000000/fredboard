package controllers

import (
	"net/http"

	"accidentallycoded.com/fredboard/v3/web/content"
	"accidentallycoded.com/fredboard/v3/web/server"
)

type staticController struct {
  *Controller
}

func NewStaticController(server *server.Web) *staticController {
  controller := &staticController{newController(server)}

  controller.mux.Handle("/", http.FileServerFS(content.ContentFS))

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
  http.ServeFile(w, r, "static/" + r.URL.Path)
}
