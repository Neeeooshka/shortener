package storage

import "net/http"

type LinkStorage interface {
	Add(string, string, string) error
	AddBatch([]Batch, string) error
	Get(string) (string, bool)
	Close() error
	PingHandler(http.ResponseWriter, *http.Request)
	GetUserURLs(string) []Link
}

type Link struct {
	UUID      uint   `json:"uuid"`
	ShortLink string `json:"short_url"`
	FullLink  string `json:"original_url"`
	UserID    string
}

type Batch struct {
	ID       string `json:"correlation_id"`
	URL      string `json:"-"`
	ShortURL string `json:"-"`
	Result   string `json:"short_url"`
}
