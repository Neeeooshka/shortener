package main

import (
	"github.com/Neeeooshka/alice-skill.git/internal/handlers"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	router := chi.NewRouter()

	router.Post("/", handlers.EndPointPOST)
	router.Get("/{id}", handlers.EndPointGET)

	log.Fatal(http.ListenAndServe(`:8080`, router))
}
