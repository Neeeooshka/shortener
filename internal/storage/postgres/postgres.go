package postgres

import (
	"context"
	"database/sql"
	"net/http"
	"sync"

	"github.com/Neeeooshka/alice-skill.git/pkg/semaphore"

	"github.com/Neeeooshka/alice-skill.git/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var wg *sync.WaitGroup
var waitWrite *sync.Once
var sem = semaphore.NewSemaphore(1)

type ConflictError struct {
	ShortLink string
}

func (e *ConflictError) Error() string {
	return "link already exsists"
}

type Postgres struct {
	DB             *sql.DB
	deleteLinksChs []chan storage.UserLinks
}

func (l *Postgres) Add(sl, fl, userID string) error {

	var shortLink string
	var isNew bool

	row := l.DB.QueryRow("WITH ins AS (\n    INSERT INTO shortener_links (short_url, original_url, user_id)\n    VALUES ($1, $2, $3)\n    ON CONFLICT (original_url) DO NOTHING\n        RETURNING short_url\n)\nSELECT short_url, 1 as is_new FROM ins\nUNION  ALL\nSELECT short_url, 0 as is_new FROM shortener_links WHERE original_url = $2\nLIMIT 1", sl, fl, userID)
	err := row.Scan(&shortLink, &isNew)
	if err != nil {
		return err
	}

	if !isNew {
		return &ConflictError{ShortLink: shortLink}
	}

	return nil
}

func (l *Postgres) AddBatch(b []storage.Batch, userID string) error {

	ctx, cansel := context.WithCancel(context.Background())
	defer cansel()

	tx, err := l.DB.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO shortener_links (short_url, original_url, user_id) VALUES ($1,$2,$3)\nON CONFLICT (original_url) DO NOTHING")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range b {
		_, err := stmt.Exec(e.ShortURL, e.URL, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (l *Postgres) Get(shortLink string) (storage.Link, bool) {

	link := storage.Link{}

	row := l.DB.QueryRow("SELECT original_url, short_url, user_id, deleted FROM shortener_links WHERE short_url = $1", shortLink)
	err := row.Scan(&link.FullLink, &link.ShortLink, &link.UserID, &link.Deleted)
	if err != nil {
		return storage.Link{}, false
	}

	return link, true
}

func (l *Postgres) GetUserURLs(userID string) []storage.Link {

	var links []storage.Link

	rows, err := l.DB.Query("SELECT short_url, original_url FROM shortener_links WHERE user_id = $1", userID)
	if err == nil && rows.Err() == nil {
		for rows.Next() {

			var shortLink, fullLink string

			err := rows.Scan(&shortLink, &fullLink)
			if err == nil {
				links = append(links, storage.Link{ShortLink: shortLink, FullLink: fullLink})
			}
		}
	}

	return links
}

func (l *Postgres) DeleteUserURLs(ul storage.UserLinks) error {

	dataCh := make(chan storage.UserLinks)
	go func() {
		waitWrite.Do(func() {
			wg = new(sync.WaitGroup)
			go func() {
				wg.Wait()
				go func() {
					sem.Acquire()
					defer sem.Release()
					for ls := range dataCh {
						_, _ = l.DB.Exec("UPDATE shortener_links SET deleted = true WHERE user_id = $1 and short_url in ($2)", ls.UserID, ls.LinksID)
					}
				}()
				waitWrite = new(sync.Once)
			}()
		})
		wg.Add(1)
		defer wg.Done()
		defer close(dataCh)
		dataCh <- ul
	}()

	// конвеер
	ch := len(l.deleteLinksChs)
	l.deleteLinksChs[ch] = make(chan storage.UserLinks)
	go func() {
		defer close(l.deleteLinksChs[ch])
		for data := range dataCh {
			l.deleteLinksChs[ch] <- data
		}
	}()

	// объединение результата fanIn
	finalCh := make(chan storage.UserLinks)

	for _, ch := range l.deleteLinksChs {
		chCopy := ch

		go func() {

			for data := range chCopy {
				finalCh <- data
			}
		}()

	}
	go func() {
		wg.Wait()
		close(finalCh)
	}()

	// обработка результата
	var wg1 sync.WaitGroup

	wg.Add(1)
	go func() {
		sem.Acquire()
		defer wg.Done()
		defer sem.Release()
		for _ = range finalCh {
			//_, _ = l.DB.Exec("UPDATE shortener_links SET deleted = true WHERE user_id = $1 and short_url in ($2)", ul.UserID, ul.LinksID)
		}
	}()
	wg1.Wait()

	return nil
}

func (l *Postgres) asda() chan int {

	resCh := make(chan int)

	go func() {

		defer close(resCh)

		for data := range l.deleteLinksChs {
			resCh <- data
		}
	}()

	return resCh
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
	_, err = l.DB.Exec("CREATE TABLE IF NOT EXISTS shortener_links (\n    id SERIAL,\n    short_url character(8) NOT NULL,\n    original_url character varying(250) NOT NULL,\n    user_id character(32) NULL,\n    deleted boolean NOT NULL DEFAULT false,\n    PRIMARY KEY (uuid),\n    UNIQUE (original_url)\n )")
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
