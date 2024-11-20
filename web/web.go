package web

import (
	"fmt"
	"net/http"
)

func Start() {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(res, "<h1>Hello, world!</h1>")
	})

	http.ListenAndServe(":8080", nil)
}
