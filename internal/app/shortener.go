package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Neeeooshka/alice-skill.git/internal/auth"

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
	linkID, _ := strings.CutPrefix(r.URL.String(), "/")
	if link, ok := a.storage.Get(linkID); ok {
		if link.Deleted {
			w.WriteHeader(http.StatusGone)
			return
		}
		w.Header().Set("Location", link.FullLink)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func (a *shortenerApp) UserUrlsHandler(w http.ResponseWriter, r *http.Request) {

	ck, err := r.Cookie("userID")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := auth.GetUserID(ck.Value)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	links := a.storage.GetUserURLs(userID)
	if len(links) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userLinks := make([]storage.Link, 0, len(links))

	for _, e := range links {
		userLinks = append(userLinks, storage.Link{ShortLink: a.GetShortURL(e.ShortLink), FullLink: e.FullLink})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&userLinks)
}

func (a *shortenerApp) DeleteUserUrlsHandler(w http.ResponseWriter, r *http.Request) {

	ck, err := r.Cookie("userID")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := auth.GetUserID(ck.Value)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	req := storage.UserLinks{UserID: userID}

	if err := json.NewDecoder(r.Body).Decode(&req.LinksID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = a.storage.DeleteUserURLs(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *shortenerApp) ShortenerHandler(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := a.getUserID(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortLink := a.GenerateShortLink()
	err = a.storage.Add(shortLink, string(body), userID)
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

	userID, err := a.getUserID(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortLink := a.GenerateShortLink()
	err = a.storage.Add(shortLink, req.URL, userID)

	w.Header().Set("Content-Type", "application/json")

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

	type reqURL struct {
		ID  string `json:"correlation_id"`
		URL string `json:"original_url"`
	}

	var req []reqURL

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := a.getUserID(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := make([]storage.Batch, 0, len(req))

	for _, e := range req {
		shortLink := a.GenerateShortLink()
		resp = append(resp, storage.Batch{ID: e.ID, URL: e.URL, ShortURL: shortLink, Result: a.GetShortURL(shortLink)})
	}

	if err := a.storage.AddBatch(resp, userID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&resp)
}

func (a *shortenerApp) getUserID(w http.ResponseWriter, r *http.Request) (string, error) {

	ck, err := r.Cookie("userID")

	if err != nil {
		token, err := auth.GenerateToken()
		if err != nil {
			return "", err
		}
		ck = &http.Cookie{Name: "userID", Value: token, Expires: time.Now().Add(time.Hour * 24 * 365), HttpOnly: true}

		http.SetCookie(w, ck)
	}

	userID, err := auth.GetUserID(ck.Value)

	if err != nil {
		return "", err
	}

	return userID, nil
}
