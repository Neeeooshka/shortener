package postgres

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"net/http"
)

type Postgres struct {
	DB *sql.DB
}

func (l *Postgres) Add(sl, fl string) error {
	_, err := l.DB.Exec("INSERT INTO shortener_links (short_url, original_url) VALUES ($1,$2)\nON CONFLICT (short_url) DO\n    UPDATE SET original_url = EXCLUDED.original_url", sl, fl)
	return err
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
	_, err = l.DB.Exec("CREATE TABLE IF NOT EXISTS shortener_links (\n    uuid SERIAL,\n    short_url character(8) NOT NULL,\n    original_url character(250) NOT NULL,\n    PRIMARY KEY (uuid),\n    UNIQUE (short_url)\n )")
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
