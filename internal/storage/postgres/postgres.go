package postgres

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/Neeeooshka/alice-skill.git/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type ConflictError struct {
	ShortLink string
	err       string
}

func (e *ConflictError) Error() string {
	return e.err
}

type Postgres struct {
	DB *sql.DB
}

func (l *Postgres) Add(sl, fl string) error {

	var shortLink string
	var isNew bool

	row := l.DB.QueryRow("WITH ins AS (\n    INSERT INTO shortener_links (short_url, original_url)\n    VALUES ($1, $2)\n    ON CONFLICT (original_url) DO NOTHING\n        RETURNING short_url\n)\nSELECT short_url, 1 as is_new FROM ins\nUNION  ALL\nSELECT short_url, 0 as is_new FROM shortener_links WHERE original_url = $2\nLIMIT 1", sl, fl)
	err := row.Scan(&shortLink, &isNew)
	if err != nil {
		return err
	}

	if !isNew {
		return &ConflictError{err: "link already exsists", ShortLink: shortLink}
	}

	return nil
}

func (l *Postgres) AddBatch(b []storage.Batch) error {

	ctx, cansel := context.WithCancel(context.Background())
	defer cansel()

	tx, err := l.DB.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO shortener_links (short_url, original_url) VALUES ($1,$2)\nON CONFLICT (original_url) DO NOTHING")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range b {
		_, err := stmt.Exec(e.ShortURL, e.URL)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (l *Postgres) Get(shortLink string) (string, bool) {

	var link string

	row := l.DB.QueryRow("SELECT original_url FROM shortener_links WHERE short_url = $1", shortLink)
	err := row.Scan(&link)
	if err != nil {
		return "", false
	}

	return link, true
}

func (l *Postgres) Close() error {
	return l.DB.Close()
}

func (l *Postgres) PingHandler(w http.ResponseWriter, r *http.Request) {
	err := l.DB.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (l *Postgres) initStructForLinks() (err error) {
	_, err = l.DB.Exec("CREATE TABLE IF NOT EXISTS shortener_links (\n    uuid SERIAL,\n    short_url character(8) NOT NULL,\n    original_url character(250) NOT NULL,\n    PRIMARY KEY (uuid),\n    UNIQUE (original_url)\n )")
	return err
}

func NewPostgresLinksStorage(conn string) (pgx *Postgres, err error) {

	pgx = &Postgres{}

	pgx.DB, err = sql.Open("pgx", conn)
	if err != nil {
		return nil, err
	}

	return pgx, pgx.initStructForLinks()
}
