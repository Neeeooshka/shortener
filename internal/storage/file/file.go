package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/Neeeooshka/alice-skill.git/internal/storage"
	"net/http"
	"os"
)

type Links struct {
	links       []storage.Link
	fileStorage *os.File
}

func (l *Links) Add(sl, fl string) (err error) {
	uuid := uint(len(l.links))
	newLink := storage.Link{ShortLink: sl, FullLink: fl, UUID: uuid + 1}
	l.links = append(l.links, newLink)

	if l.fileStorage != nil {
		err = json.NewEncoder(l.fileStorage).Encode(newLink)
	}

	return err
}

func (l *Links) Get(shortLink string) (string, bool) {
	for _, link := range l.links {
		if link.ShortLink == shortLink {
			return link.FullLink, true
		}
	}
	return "", false
}

func (l *Links) Close() error {
	return l.fileStorage.Close()
}

func (l *Links) PingHandler(w http.ResponseWriter, r *http.Request) {

	if l.fileStorage == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (l *Links) SetLinksFromFile(filename string) error {
	if filename == "" {
		return errors.New("the param filename is not set")
	}

	file, err := os.Open(filename)

	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		link := storage.Link{}
		if err := json.Unmarshal(scanner.Bytes(), &link); err != nil {
			continue
		}
		l.links = append(l.links, link)
	}

	if len(l.links) == 0 {
		return errors.New("links were not imported from file")
	}

	return nil
}

func NewFileLinksStorage(filename string) (links *Links, err error) {

	links = &Links{}

	if filename != "" {
		_ = links.SetLinksFromFile(filename)
		links.fileStorage, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	}

	return links, err
}
