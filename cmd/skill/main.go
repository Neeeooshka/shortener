// пакеты исполняемых приложений должны называться main
package main

import (
	"github.com/Neeeooshka/alice-skill.git/internal/handlers"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	router := chi.NewRouter()
	router.Post("/", handlers.AliceSkill)
	router.Get("/", handlers.AliceSkill)

	log.Fatal(http.ListenAndServe(`:8080`, router))
}
