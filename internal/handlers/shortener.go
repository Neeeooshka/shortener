package handlers

import (
	"fmt"
	"github.com/Neeeooshka/alice-skill.git/internal/storage"
	"github.com/thanhpk/randstr"
	"io"
	"net/http"
	"strings"
)

const hostURL = "http://localhost:8080"

var shortedLinks storage.Links

// обработчик HTTP-запроса GET
func EndPointGET(w http.ResponseWriter, r *http.Request) {
	if idLink, ok := strings.CutPrefix(r.URL.Path, "/"); ok {
		if fullLink, ok := shortedLinks.Get(idLink); ok {
			w.Header().Set("Location", fullLink)
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
	}

	w.WriteHeader(http.StatusBadRequest)
}

// обработчик HTTP-запроса POST
func EndPointPOST(w http.ResponseWriter, r *http.Request) {

	if !strings.Contains(r.Header.Get("Content-Type"), "text/plain") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idLink := generateShortLink()
	shortedLinks.Add(idLink, string(body))

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, hostURL+"/"+idLink)
}

// generate a short link
func generateShortLink() string {
	return randstr.Base62(8)
}
