package main

import (
	"net/http"
	"regexp"
)

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
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	switch r.Method {
	case http.MethodPost:
		endPointPOST(w, r)
	case http.MethodGet:
		endPointGET(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

// обработчик HTTP-запроса POST
func endPointPOST(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`method POST`))
}

// обработчик HTTP-запроса GET
func endPointGET(w http.ResponseWriter, r *http.Request) {

	if match, err := regexp.MatchString("[A-Za-z0-9]", r.URL.String()); err != nil || !match {
		http.Error(w, "link is incorrect", http.StatusBadRequest)
	}

	w.Write([]byte(`method GET`))
}
