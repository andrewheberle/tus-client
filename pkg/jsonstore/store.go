package jsonstore

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/eventials/go-tus"
)

// JsonStore implements a JSON file based store for resumable uploads via github.com/eventials/go-tus
//
// Locking is used to ensure concurrent access within the same JsonStore is safe, however using the
// same underlying JSON file for more than one JsonStore may cause data loss or corruption.
type JsonStore struct {
	file  string
	store map[string]string
	mu    sync.RWMutex
}

// NewJsonStore creates a new JsonStore using the provided file path
func NewJsonStore(file string) (tus.Store, error) {
	f, err := os.Open(file)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			// some other error so fail here
			return nil, err
		}

		// store doesn't exist return an empty one
		return &JsonStore{file: file, store: make(map[string]string)}, nil
	}
	defer f.Close()

	// try to load if it existed
	var store map[string]string
	dec := json.NewDecoder(f)
	if err := dec.Decode(&store); err != nil {
		return nil, err
	}

	return &JsonStore{file: file, store: store}, nil
}

func (s *JsonStore) Get(fingerprint string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.store[fingerprint]

	return url, ok
}

func (s *JsonStore) Set(fingerprint, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[fingerprint] = url

	s.savestore()
}

func (s *JsonStore) Delete(fingerprint string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.store, fingerprint)

	s.savestore()
}

func (s *JsonStore) Close() {
	s.savestore()
}

func (s *JsonStore) savestore() error {
	f, err := os.CreateTemp(filepath.Dir(s.file), filepath.Base(s.file)+"*")
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.Remove(f.Name())

	enc := json.NewEncoder(f)
	if err := enc.Encode(s.store); err != nil {
		return err
	}
	f.Close()

	if err := os.Rename(f.Name(), s.file); err != nil {
		return err
	}

	return nil
}
