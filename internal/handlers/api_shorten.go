package handlers

import (
	"encoding/json"
	"fmt"
	"io"
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
			fmt.Println(io.ReadAll(r.Body))
			fmt.Println(req)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		shortLink := s.GenerateShortLink()
		s.Add(shortLink, req.URL)

		resp := struct {
			Result string `json:"result"`
		}{
			Result: s.GetBaseURL() + "/" + shortLink,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(&resp)
	}
}
