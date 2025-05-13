package storage

import (
	"context"
	"net/http"
)

type LinkStorage interface {
	Add(string, string, string) error
	AddBatch(context.Context, []Batch, string) error
	Get(string) (Link, bool)
	Close() error
	PingHandler(http.ResponseWriter, *http.Request)
	GetUserURLs(string) []Link
	DeleteUserURLs([]UserLinks) error
}

type Link struct {
	UserID    string `json:"uuid,omitempty" db:"user_id"`
	ShortLink string `json:"short_url" db:"short_url"`
	FullLink  string `json:"original_url" db:"original_url"`
	Deleted   bool   `json:"deleted" db:"deleted"`
}

type Batch struct {
	ID       string `json:"correlation_id"`
	URL      string `json:"-"`
	ShortURL string `json:"-"`
	Result   string `json:"short_url"`
}

type UserLinks struct {
	LinksID []string
	UserID  string
}
