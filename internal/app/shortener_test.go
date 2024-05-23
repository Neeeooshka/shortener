package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Neeeooshka/alice-skill.git/internal/config"
	storage "github.com/Neeeooshka/alice-skill.git/internal/storage/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type link struct {
	URL         string `json:"url"`
	ShortedLink string `json:"result"`
}

var testCases = []struct {
	link link
}{
	{
		link: link{
			URL: "https://ya.ru",
		},
	},
	{
		link: link{
			URL: "https://google.com",
		},
	},
	{
		link: link{
			URL: "https://practicum.yandex.ru",
		},
	},
}

func TestShortener(t *testing.T) {

	app := &shortenerApp{Options: config.NewOptions()}

	for _, tc := range testCases {
		t.Run("shorte link: "+tc.link.URL, func(t *testing.T) {
			// get shorted link
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.link.URL))
			w := httptest.NewRecorder()

			r.Header.Set("Content-Type", "text/plain")

			app.setMockShortLink(t, tc.link.URL)
			app.ShortenerHandler(w, r)

			require.Equal(t, http.StatusCreated, w.Code)

			tc.link.ShortedLink = w.Body.String()

			u, err := url.Parse(tc.link.ShortedLink)
			require.NoError(t, err)

			// get full link from shorted
			r = httptest.NewRequest(http.MethodGet, u.RequestURI(), nil)
			w = httptest.NewRecorder()

			app.ExpanderHandler(w, r)

			require.Equal(t, http.StatusTemporaryRedirect, w.Code)
			assert.Equal(t, w.Header().Get("Location"), tc.link.URL)
		})
	}
}

func TestApiShorten(t *testing.T) {

	app := &shortenerApp{Options: config.NewOptions()}

	for _, tc := range testCases {
		t.Run("shorte link: "+tc.link.URL, func(t *testing.T) {
			// get shorted link
			body := bytes.NewBuffer([]byte(""))
			require.NoError(t, json.NewEncoder(body).Encode(&tc.link))

			r := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
			w := httptest.NewRecorder()

			r.Header.Set("Content-Type", "application/json")

			app.setMockShortLink(t, tc.link.URL)
			app.APIShortenerHandler(w, r)

			require.Equal(t, http.StatusCreated, w.Code)
			require.NoError(t, json.NewDecoder(w.Body).Decode(&tc.link))

			// get full link from shorted
			u, err := url.Parse(tc.link.ShortedLink)
			require.NoError(t, err)

			r = httptest.NewRequest(http.MethodGet, u.RequestURI(), nil)
			w = httptest.NewRecorder()

			app.ExpanderHandler(w, r)

			require.Equal(t, http.StatusTemporaryRedirect, w.Code)
			assert.Equal(t, w.Header().Get("Location"), tc.link.URL)
		})
	}
}

func (a *shortenerApp) setMockShortLink(t *testing.T, URL string) {

	ctrl := gomock.NewController(t)
	s := storage.NewMockLinkStorage(ctrl)

	s.EXPECT().Add(gomock.Any(), URL, gomock.Any()).Return(nil)
	s.EXPECT().Get(gomock.Any()).Return(URL, true)

	a.storage = s
}
