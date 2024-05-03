package storage

type LinkStorage interface {
	Add(sl, fl string) error
	Get(shortLink string) (string, bool)
}

type Link struct {
	UUID      uint   `json:"uuid"`
	ShortLink string `json:"short_url"`
	FullLink  string `json:"original_url"`
}
