// boltstore implements a [go.etcd.io/bbolt] based store for resumable uploads via [github.com/eventials/go-tus]
package boltstore

import (
	"github.com/andrewheberle/ubolt"
	tus "github.com/eventials/go-tus"
)

// BoltStore implements the [tus.Store] interface
type BoltStore struct {
	db *ubolt.Bucket
}

// NewBoltStore creates a new BoltStore using the provided file path
func NewBoltStore(database string) (tus.Store, error) {
	db, err := ubolt.OpenBucket(database, []byte("resume"))
	if err != nil {
		return nil, err
	}

	return &BoltStore{db}, nil
}

func (s *BoltStore) Get(fingerprint string) (string, bool) {
	url, err := s.db.GetE([]byte(fingerprint))
	if err != nil {
		return "", false
	}

	return string(url), true
}

func (s *BoltStore) Set(fingerprint, url string) {
	s.db.Put([]byte(fingerprint), []byte(url))
}

func (s *BoltStore) Delete(fingerprint string) {
	s.db.Delete([]byte(fingerprint))
}

func (s *BoltStore) Close() {
	s.db.Close()
}
