package server

import (
	"net/http"
)

type Handler struct {
	Route   string
	Handler func(http.ResponseWriter, *http.Request)
}

// run HTTP server
func RunHTTPServer(handlers []Handler) error {
	serverMux := http.NewServeMux()
	for _, h := range handlers {
		serverMux.HandleFunc(h.Route, h.Handler)
	}

	return http.ListenAndServe(`:8080`, serverMux)
}
