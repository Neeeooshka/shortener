package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/Neeeooshka/alice-skill.git/internal/storage"
)

type Links struct {
	links       []storage.Link
	fileStorage *os.File
}

func (l *Links) Add(sl, fl, userID string) (err error) {

	newLink := storage.Link{UserID: userID, ShortLink: sl, FullLink: fl, Deleted: false}
	l.links = append(l.links, newLink)

	if l.fileStorage != nil {
		err := json.NewEncoder(l.fileStorage).Encode(newLink)
		if err != nil {
			return err
		}
	}

	return err
}

func (l *Links) AddBatch(ctx context.Context, b []storage.Batch, userID string) error {
	for _, e := range b {
		err := l.Add(e.ShortURL, e.URL, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Links) Get(shortLink string) (storage.Link, bool) {
	for _, link := range l.links {
		if link.ShortLink == shortLink {
			return link, true
		}
	}
	return storage.Link{}, false
}

func (l *Links) Close() error {
	return l.fileStorage.Close()
}

func (l *Links) PingHandler(w http.ResponseWriter, _ *http.Request) {

	if l.fileStorage == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (l *Links) GetUserURLs(userID string) []storage.Link {

	var links = make([]storage.Link, 0, len(l.links))

	for _, link := range l.links {
		if link.UserID == userID {
			links = append(links, link)
		}
	}

	return links
}

func (l *Links) DeleteUserURLs(uls []storage.UserLinks) error {

	ulMap := make(map[string]string)

	for _, ul := range uls {
		for _, shortLink := range ul.LinksID {
			ulMap[shortLink] = ul.UserID
		}
	}

	for i, link := range l.links {

		userID, ok := ulMap[link.ShortLink]

		if ok && userID == link.UserID {
			l.links[i].Deleted = true
		}
	}

	if l.fileStorage != nil {
		_, err := l.fileStorage.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		err = l.fileStorage.Truncate(0)
		if err != nil {
			return err
		}

		for _, link := range l.links {
			err := json.NewEncoder(l.fileStorage).Encode(link)
			if err != nil {
				return err
			}
		}
	}

	return nil
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
