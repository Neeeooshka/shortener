package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Neeeooshka/alice-skill.git/internal/config"
	"github.com/Neeeooshka/alice-skill.git/internal/storage"
	"github.com/Neeeooshka/alice-skill.git/internal/storage/postgres"
	"github.com/thanhpk/randstr"
)

type shortenerApp struct {
	Options config.Options
	storage storage.LinkStorage
}

func (a *shortenerApp) GetShortURL(id string) string {
	return a.Options.GetBaseURL() + "/" + id
}

func (a *shortenerApp) GenerateShortLink() string {
	return randstr.Base62(8)
}

func NewShortenerAppInstance(opt config.Options, s storage.LinkStorage) *shortenerApp {
	return &shortenerApp{Options: opt, storage: s}
}

func (a *shortenerApp) ExpanderHandler(w http.ResponseWriter, r *http.Request) {
	link, _ := strings.CutPrefix(r.URL.String(), "/")
	if fullLink, ok := a.storage.Get(link); ok {
		w.Header().Set("Location", fullLink)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func (a *shortenerApp) ShortenerHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortLink := a.GenerateShortLink()
	err = a.storage.Add(shortLink, string(body))
	var ce *postgres.ConflictError
	if err != nil {
		if errors.As(err, &ce) {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, a.GetShortURL(ce.ShortLink))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, a.GetShortURL(shortLink))
}

func (a *shortenerApp) APIShortenerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortLink := a.GenerateShortLink()
	err := a.storage.Add(shortLink, req.URL)

	w.Header().Set("Content-Type", "ShortenerApplication/json")

	resp := struct {
		Result string `json:"result"`
	}{
		Result: a.GetShortURL(shortLink),
	}

	var ce *postgres.ConflictError
	if err != nil {
		if !errors.As(err, &ce) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusConflict)
		resp.Result = a.GetShortURL(ce.ShortLink)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	json.NewEncoder(w).Encode(&resp)
}

func (a *shortenerApp) APIBatchShortenerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type reqURL struct {
		ID  string `json:"correlation_id"`
		URL string `json:"original_url"`
	}

	var req []reqURL

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp := make([]storage.Batch, 0, len(req))

	for _, e := range req {
		shortLink := a.GenerateShortLink()
		resp = append(resp, storage.Batch{ID: e.ID, URL: e.URL, ShortURL: shortLink, Result: a.GetShortURL(shortLink)})
	}

	if err := a.storage.AddBatch(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "ShortenerApplication/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&resp)
}
