package handlers

import (
	"fmt"
	"github.com/Neeeooshka/alice-skill.git/cmd/config"
	"github.com/Neeeooshka/alice-skill.git/internal/storage"
	"github.com/thanhpk/randstr"
	"io"
	"net/http"
	"strings"
)

var shortedLinks storage.Links

// обработчик HTTP-запроса GET
func EndPointGET(w http.ResponseWriter, r *http.Request) {

	opt := config.GetOptions()
	rLink := opt.GetShortLinkServer() + r.URL.String()
	if fullLink, ok := shortedLinks.Get(rLink); ok {
		w.Header().Set("Location", fullLink)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
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

	shortLink := generateShortLink()
	shortedLinks.Add(shortLink, string(body))

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortLink)
}

// generate a short link
func generateShortLink() string {
	opt := config.GetOptions()
	return opt.GetShortLinkServer() + "/" + randstr.Base62(8)
}
