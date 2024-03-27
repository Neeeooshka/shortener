package main

import (
	"flag"
	"github.com/Neeeooshka/alice-skill.git/cmd/config"
	"github.com/Neeeooshka/alice-skill.git/internal/handlers"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	router := chi.NewRouter()

	router.Post("/", handlers.EndPointPOST)
	router.Get("/{id}", handlers.EndPointGET)

	opt := config.GetOptions()

	var a, b string

	flag.StringVar(&a, "a", opt.GetServer(), "Server address host:port")
	flag.StringVar(&b, "b", opt.GetShortLinkServer(), "Server ShortLink address protocol://host:port")
	flag.Parse()

	opt.SetServer(a)
	opt.SetShortLinkServer(b)

	log.Fatal(http.ListenAndServe(opt.GetServer(), router))
}
