// пакеты исполняемых приложений должны называться main
package main

import (
	"fmt"
	"github.com/Neeeooshka/alice-skill.git/internal/handlers"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	// обрабатываем аргументы командной строки
	parseFlags()

	router := chi.NewRouter()
	router.Post("/", handlers.AliceSkill)
	router.Get("/", handlers.AliceSkill)

	fmt.Println("Running server on", flagRunAddr)
	log.Fatal(http.ListenAndServe(flagRunAddr, router))
}
