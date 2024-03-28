package main

import (
	"flag"
	"github.com/Neeeooshka/alice-skill.git/internal/config"
	"github.com/Neeeooshka/alice-skill.git/internal/handlers"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {

	flag.Parse()

	opt := config.GetOptions()
	cnf := config.GetConfig()

	router := chi.NewRouter()
	router.Post("/", handlers.EndPointPOST)
	router.Get("/{id}", handlers.EndPointGET)

	server := opt.GetServer()
	if cnf.ServerAddress != "" {
		server = cnf.ServerAddress
	}
	log.Fatal(http.ListenAndServe(server, router))
}
