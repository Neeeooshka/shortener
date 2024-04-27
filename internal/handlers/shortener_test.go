package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

const serverAddress = "localhost"

func TestShortener(t *testing.T) {

	testCases := []struct {
		link string
		id   string
		port int
	}{
		{
			link: "https://ya.ru",
			id:   "JdHkDaPe",
			port: 8080,
		},
		{
			link: "https://google.com",
			id:   "uErYlAmX",
			port: 8888,
		},
	}

	for _, tc := range testCases {

		var shortedLink string

		s := newShortener(tc.port, tc.id)

		// get shorted link
		t.Run("shorte link: "+tc.link, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.link))
			w := httptest.NewRecorder()

			r.Header.Set("Content-Type", "text/plain")

			GetShortenerHandler(&s)(w, r)

			require.Equal(t, http.StatusCreated, w.Code)

			shortedLink = w.Body.String()
		})
		// get full link from shorted
		t.Run("expand link: "+tc.link, func(t *testing.T) {
			u, err := url.Parse(shortedLink)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodGet, u.RequestURI(), strings.NewReader(""))
			w := httptest.NewRecorder()

			GetExpanderHandler(&s)(w, r)

			require.Equal(t, http.StatusTemporaryRedirect, w.Code)
			assert.Equal(t, w.Header().Get("Location"), tc.link)
		})
	}
}

type shortener struct {
	port int
	id   string
	fl   string
}

func (s *shortener) GetBaseURL() string {
	return "http://" + serverAddress + ":" + strconv.Itoa(s.port)
}

func (s *shortener) GenerateShortLink() string {
	return s.id
}

func (s *shortener) Add(sl, fl string) {
	s.fl = fl
	s.id = sl
}
func (s *shortener) Get(sl string) (string, bool) {
	return s.fl, true
}

func newShortener(port int, linkID string) shortener {
	return shortener{port: port, id: linkID}
}
