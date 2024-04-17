package handlers

import (
	"fmt"
	"github.com/Neeeooshka/alice-skill.git/internal/storage"
	"io"
	"net/http"
)

type Shortener interface {
	GetBaseURL() string
	GenerateShortLink() string
}

var shortedLinks storage.Links

// Возвращает обработчик HTTP-запроса GET
func GetExpanderHandler(s Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rLink := s.GetBaseURL() + r.URL.String()
		if fullLink, ok := shortedLinks.Get(rLink); ok {
			w.Header().Set("Location", fullLink)
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	}
}

// Возвращает обработчик HTTP-запроса POST
func GetShortenerHandler(s Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		shortLink := s.GenerateShortLink()
		shortedLinks.Add(shortLink, string(body))

		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "text/plain")
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, shortLink)
	}
}
