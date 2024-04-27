package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type link struct {
	Uuid      uint   `json:"uuid"`
	ShortLink string `json:"short_url"`
	FullLink  string `json:"original_url"`
}

type Links struct {
	links       []link
	fileStorage *os.File
	encoder     *json.Encoder
	useFile     bool
}

func (l *Links) Add(sl, fl string) {
	uuid := uint(len(l.links))
	newLink := link{ShortLink: sl, FullLink: fl, Uuid: uuid + 1}
	l.links = append(l.links, newLink)

	if l.useFile {
		err := l.encoder.Encode(newLink)
		_ = err
	}
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
	if !l.useFile {
		return errors.New("file storage is not include")
	}

	scanner := bufio.NewScanner(l.fileStorage)
	for scanner.Scan() {
		link := link{}
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

func NewLinksStorage(filename string) *Links {

	var links = &Links{}

	if filename != "" {
		dir, filename := filepath.Split(filename)
		if len(dir) > 0 && os.IsPathSeparator(dir[0]) {
			dir, _ = strings.CutPrefix(dir, string(dir[0]))
			os.MkdirAll(dir, 0666)
			filename = dir + filename
		}
		file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

		if err == nil {
			//defer file.Close()
			links.fileStorage = file
			links.useFile = true
			links.encoder = json.NewEncoder(links.fileStorage)
			links.SetLinksFromFile()
		}
	}

	return links
}
