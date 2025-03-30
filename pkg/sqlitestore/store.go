package sqlitestore

import (
	"database/sql"

	"github.com/eventials/go-tus"

	_ "github.com/glebarez/go-sqlite"
)

// SqliteStore implements an SQLite based store for resumable uploads via github.com/eventials/go-tus
type SqliteStore struct {
	db *sql.DB
}

// NewSqliteStore creates a new SqliteStore using the provided database path
func NewSqliteStore(database string) (tus.Store, error) {
	db, err := sql.Open("sqlite", database)
	if err != nil {
		return nil, err
	}

	// create table if it didn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS fingerprints(fingerprint TEXT PRIMARY KEY, url TEXT)")
	if err != nil {
		return nil, err
	}

	return &SqliteStore{db}, nil
}

func (s *SqliteStore) Get(fingerprint string) (string, bool) {
	var url string
	err := s.db.QueryRow("SELECT url FROM fingerprints WHERE fingerprint=?", fingerprint).Scan(&url)
	if err != nil {
		return "", false
	}

	return url, true
}

func (s *SqliteStore) Set(fingerprint, url string) {
	s.db.Exec("INSERT INTO fingerprints(fingerprint, url) VALUES(?, ?) ON CONFLICT DO UPDATE SET url=excluded.url", fingerprint, url)
}

func (s *SqliteStore) Delete(fingerprint string) {
	s.db.Exec("DELETE FROM fingerprints WHERE fingerprint=?", fingerprint)
}

func (s *SqliteStore) Close() {
	s.db.Close()
}
