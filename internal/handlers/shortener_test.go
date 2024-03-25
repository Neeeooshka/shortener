package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortener(t *testing.T) {

	testCases := []struct {
		link string
	}{
		{link: "https://ya.ru"},
		{link: "https://google.com"},
	}

	for _, tc := range testCases {

		var shortedLink string

		// get shorted link
		t.Run("shorte link: "+tc.link, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.link))
			w := httptest.NewRecorder()

			r.Header.Set("Content-Type", "text/plain")

			EndPointPOST(w, r)

			require.Equal(t, http.StatusCreated, w.Code)

			shortedLink = w.Body.String()
		})
		// get full link from shorted
		t.Run("expand link: "+shortedLink, func(t *testing.T) {

			r := httptest.NewRequest(http.MethodGet, shortedLink, nil)
			w := httptest.NewRecorder()

			EndPointGET(w, r)

			require.Equal(t, http.StatusTemporaryRedirect, w.Code)
			assert.Equal(t, w.Header().Get("Location"), tc.link)
		})
	}
}
