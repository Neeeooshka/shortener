package handlers

import (
	"fmt"
	"github.com/Neeeooshka/alice-skill.git/internal/config"
	"github.com/Neeeooshka/alice-skill.git/internal/storage"
	"github.com/thanhpk/randstr"
	"io"
	"net/http"
	"strings"
)

var shortedLinks storage.Links

func getBaseURL() string {
	opt := config.GetOptions()
	cnf := config.GetConfig()

	baseURL := opt.GetBaseURL()
	if cnf.BaseURL != "" {
		baseURL = cnf.BaseURL
	}

	return baseURL
}

// обработчик HTTP-запроса GET
func EndPointGET(w http.ResponseWriter, r *http.Request) {

	rLink := getBaseURL() + r.URL.String()
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
	return getBaseURL() + "/" + randstr.Base62(8)
}
