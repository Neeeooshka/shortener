package handlers

import (
	"encoding/json"
	"net/http"
)

// Возвращает обработчик HTTP-запроса POST для REST API
func GetAPIShortenHandler(s Shortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		shortLink := s.GenerateShortLink()
		shortedLinks.Add(shortLink, req.URL)

		resp := struct {
			Result string `json:"result"`
		}{
			Result: shortLink,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(&resp)
	}
}
