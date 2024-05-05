package storage

import (
	"database/sql"
	_ "github.com/jackc/pgx"
	"net/http"
)

type Postgres struct {
	db *sql.DB
}

func (l *Postgres) Add(sl, fl string) error {
	return nil
}

func (l *Postgres) Get(shortLink string) (string, bool) {
	return "", false
}

func (l *Postgres) PingHandler(w http.ResponseWriter, r *http.Request) {
	err := l.db.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func NewPostgresLinksStorage(conn string) (*Postgres, error) {
	db, err := sql.Open("pgx", conn)
	if err != nil {
		return &Postgres{}, err
	}
	defer db.Close()

	return &Postgres{db: db}, nil
}
