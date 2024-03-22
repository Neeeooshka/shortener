package main

import (
	"github.com/Neeeooshka/alice-skill.git/internal/handlers"
	"github.com/Neeeooshka/alice-skill.git/internal/server"
	"log"
	"net/http"
)

func main() {

	var sh []server.Handler

	handlerShortener := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.EndPointPOST(w, r)
		case http.MethodGet:
			handlers.EndPointGET(w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	sh = append(sh, server.Handler{Route: "/", Handler: handlerShortener})
	log.Fatal(server.RunHTTPServer(sh))
}
