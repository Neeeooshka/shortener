package main

import (
	"fmt"
	"github.com/thanhpk/randstr"
	"io"
	"net/http"
	"strings"
)

const hostUrl = "http://localhost:8080"

type link struct {
	shortLink string
	fullLink  string
}

type links []link

func (l *links) addLink(sl, fl string) {
	*l = append(*l, link{shortLink: sl, fullLink: fl})
}

var shortedLinks links

func main() {
	if err := runHTTPServer(); err != nil {
		panic(err)
	}
}

// run HTTP server
func runHTTPServer() error {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", HTTPHandler)
	err := http.ListenAndServe(`:8080`, serverMux)
	return err
}

func HTTPHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		endPointPOST(w, r)
	case http.MethodGet:
		endPointGET(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// обработчик HTTP-запроса POST
func endPointPOST(w http.ResponseWriter, r *http.Request) {

	if !strings.Contains(r.Header.Get("Content-Type"), "text/plain") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, generateLink(string(body)))
}

// обработчик HTTP-запроса GET
func endPointGET(w http.ResponseWriter, r *http.Request) {

	rLink := hostUrl + r.URL.String()
	if fullLink, ok := getFullLink(rLink); ok {
		w.Header().Set("Location", fullLink)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func getFullLink(shortLink string) (string, bool) {
	for _, l := range shortedLinks {
		if l.shortLink == shortLink {
			return l.fullLink, true
		}
	}
	return "", false
}

// generate a short link
func generateLink(fullLink string) string {
	shortLink := hostUrl + "/" + randstr.Base62(8)
	shortedLinks.addLink(shortLink, fullLink)

	return shortLink
}
