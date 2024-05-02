package handlers

import (
	"fmt"
	"github.com/Neeeooshka/alice-skill.git/internal/storage"
	"io"
	"net/http"
	"strings"
)

type Shortener interface {
	GetBaseURL() string
	GenerateShortLink() string
	Add(string, string)
	Get(string) (string, bool)
}

var shortedLinks storage.Links

// Возвращает обработчик HTTP-запроса GET
func GetExpanderHandler(s Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		link, _ := strings.CutPrefix(r.URL.String(), "/")
		if fullLink, ok := s.Get(link); ok {
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
		s.Add(shortLink, string(body))

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, s.GetBaseURL()+"/"+shortLink)
	}
}
