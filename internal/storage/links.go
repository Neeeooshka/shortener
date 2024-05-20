package storage

import "net/http"

type LinkStorage interface {
	Add(sl, fl string) error
	AddBatch(b []Batch) error
	Get(shortLink string) (string, bool)
	Close() error
	PingHandler(http.ResponseWriter, *http.Request)
}

type Link struct {
	UUID      uint   `json:"uuid"`
	ShortLink string `json:"short_url"`
	FullLink  string `json:"original_url"`
}

type Batch struct {
	ID       string `json:"correlation_id"`
	URL      string `json:"-"`
	ShortURL string `json:"-"`
	Result   string `json:"short_url"`
}
