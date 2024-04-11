package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestApiShorten(t *testing.T) {

	type link struct {
		URL         string `json:"url"`
		ShortedLink string `json:"result"`
	}

	testCases := []struct {
		link link
		id   string
		port int
	}{
		{
			link: link{
				URL: "https://ya.ru",
			},
			id:   "JdHkDaPe",
			port: 8080,
		},
		{
			link: link{
				URL: "https://google.com",
			},
			id:   "uErYlAmX",
			port: 8888,
		},
		{
			link: link{
				URL: "https://practicum.yandex.ru",
			},
			id:   "lypxUyCp",
			port: 8888,
		},
	}

	for _, tc := range testCases {

		s := newShortener(tc.port, tc.id)

		// get shorted link
		t.Run("shorte link: "+tc.link.URL, func(t *testing.T) {

			body := bytes.NewBuffer([]byte(""))
			require.NoError(t, json.NewEncoder(body).Encode(&tc.link))

			r := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
			w := httptest.NewRecorder()

			r.Header.Set("Content-Type", "application/json")

			GetAPIShortenHandler(&s)(w, r)

			require.Equal(t, http.StatusCreated, w.Code)
			require.NoError(t, json.NewDecoder(w.Body).Decode(&tc.link))
		})
		// get full link from shorted
		t.Run("expand link: "+tc.link.ShortedLink, func(t *testing.T) {
			u, err := url.Parse(tc.link.ShortedLink)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodGet, u.RequestURI(), nil)
			w := httptest.NewRecorder()

			GetExpanderHandler(&s)(w, r)

			require.Equal(t, http.StatusTemporaryRedirect, w.Code)
			assert.Equal(t, w.Header().Get("Location"), tc.link.URL)
		})
	}
}
