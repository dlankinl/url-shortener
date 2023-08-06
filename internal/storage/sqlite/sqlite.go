package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"url-shortener/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const fn = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	query, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
	    id INTEGER PRIMARY KEY,
	    alias TEXT NOT NULL UNIQUE,
	    url TEXT NOT NULL,
	    user TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	_, err = query.Exec()

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(alias string, urlToSave string, user string) (int64, error) {
	const fn = "storage.sqlite.SaveURL"

	query, err := s.db.Prepare("INSERT INTO url(url, alias, user) VALUES(?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s while prepairing: %w", fn, err)
	}

	res, err := query.Exec(urlToSave, alias, user)
	if err != nil {
		if sqliteError, ok := err.(sqlite3.Error); ok && sqliteError.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s while executing: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s while getting id: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const fn = "storage.sqlite.GetURL"

	query, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s while prepairing: %w", fn, err)
	}

	var urlToFind string
	err = query.QueryRow(alias).Scan(&urlToFind)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s while executing: %w", fn, err)
	}

	return urlToFind, nil
}

func (s *Storage) DeleteAlias(alias string, username string) error {
	const fn = "storage.sqlite.DeleteAlias"

	query, err := s.db.Prepare("DELETE FROM url WHERE alias = ? AND user = ?")
	if err != nil {
		return fmt.Errorf("%s while prepairing: %w", fn, err)
	}

	_, err = query.Exec(alias, username)
	if err != nil {
		return fmt.Errorf("%s while executing: %w", fn, err)
	}

	return nil
}
