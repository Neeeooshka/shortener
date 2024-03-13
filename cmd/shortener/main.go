package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type link struct {
	shortLink string
	fullLink  string
}

type links []link

func (l *links) addLink(sl, fl string) {
	*l = append(*l, link{shortLink: sl, fullLink: fl})
}

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

	if r.Header.Get("Content-Type") != "text/plain" && r.Header.Get("Content-Type") != "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

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

	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if shortLink, err := generateLink(string(body)); err == nil {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, shortLink)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

// обработчик HTTP-запроса GET
func endPointGET(w http.ResponseWriter, r *http.Request) {

	rLink := r.URL.String()

	if match, err := regexp.MatchString("[A-Za-z0-9]", rLink); rLink != "/" && (err != nil || !match) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var links = getLinks()

	for _, l := range links {
		if l.shortLink == rLink {
			w.Header().Set("Location", l.fullLink)
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
	}

	if r.URL.String() != "/" {
		w.WriteHeader(http.StatusBadRequest)
	}
}

// имитация запроса к БД
func getLinks() links {
	var l links
	l.addLink("/EwHXdJfB", "https://practicum.yandex.ru/")

	return l
}

// generate a short link
func generateLink(shortLink string) (string, error) {

	// имитация генерации ссылки
	if shortLink == "https://practicum.yandex.ru/" {
		return "EwHXdJfB", nil
	}

	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}
