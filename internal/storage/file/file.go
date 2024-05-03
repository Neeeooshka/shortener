package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/Neeeooshka/alice-skill.git/internal/storage"
	"os"
)

type Links struct {
	links       []storage.Link
	fileStorage string
}

func (l *Links) Add(sl, fl string) error {
	uuid := uint(len(l.links))
	newLink := storage.Link{ShortLink: sl, FullLink: fl, UUID: uuid + 1}
	l.links = append(l.links, newLink)

	if l.fileStorage != "" {
		file, err := os.OpenFile(l.fileStorage, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}

		defer file.Close()
		json.NewEncoder(file).Encode(newLink)
	}

	return nil
}

func (l *Links) Get(shortLink string) (string, bool) {
	for _, link := range l.links {
		if link.ShortLink == shortLink {
			return link.FullLink, true
		}
	}
	return "", false
}

func (l *Links) SetLinksFromFile() error {
	if l.fileStorage == "" {
		return errors.New("file storage is not include")
	}

	file, err := os.Open(l.fileStorage)

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

func NewFileLinksStorage(filename string) *Links {

	var links = &Links{}

	links.fileStorage = filename
	links.SetLinksFromFile()

	return links
}
