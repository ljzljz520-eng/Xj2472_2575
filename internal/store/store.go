package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	bolt "go.etcd.io/bbolt"

	"coldchain-alert/internal/domain"
)

const (
	BucketZones    = "zones"
	BucketDevices  = "devices"
	BucketReadings = "temperature_readings"
	BucketAlerts   = "temperature_alerts"
	BucketReviews  = "alert_reviews"
	BucketDoors    = "door_events"
	BucketAudits   = "audit_entries"
)

var buckets = [][]byte{
	[]byte(BucketZones),
	[]byte(BucketDevices),
	[]byte(BucketReadings),
	[]byte(BucketAlerts),
	[]byte(BucketReviews),
	[]byte(BucketDoors),
	[]byte(BucketAudits),
}

type Store struct {
	db   *bolt.DB
	mu   sync.RWMutex
	path string
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("%w: database path is required", domain.ErrInvalidInput)
	}
	directory := filepath.Dir(path)
	if directory != "." {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}
	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("open bbolt database: %w", err)
	}
	storage := &Store{db: db, path: path}
	if err := storage.initialize(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return storage, nil
}

func (s *Store) initialize() error {
	if s == nil || s.db == nil {
		return domain.ErrStorageClosed
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return fmt.Errorf("create bucket %s: %w", string(bucket), err)
			}
		}
		return nil
	})
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) Path() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.path
}

func (s *Store) withView(fn func(*bolt.Tx) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return domain.ErrStorageClosed
	}
	return s.db.View(fn)
}

func (s *Store) withUpdate(fn func(*bolt.Tx) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return domain.ErrStorageClosed
	}
	return s.db.Update(fn)
}

func putJSON(tx *bolt.Tx, bucket, key string, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode %s/%s: %w", bucket, key, err)
	}
	container := tx.Bucket([]byte(bucket))
	if container == nil {
		return fmt.Errorf("bucket %s is missing", bucket)
	}
	return container.Put([]byte(key), encoded)
}

func getJSON(tx *bolt.Tx, bucket, key string, target any) error {
	container := tx.Bucket([]byte(bucket))
	if container == nil {
		return fmt.Errorf("bucket %s is missing", bucket)
	}
	data := container.Get([]byte(key))
	if len(data) == 0 {
		return domain.ErrNotFound
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode %s/%s: %w", bucket, key, err)
	}
	return nil
}

func deleteValue(tx *bolt.Tx, bucket, key string) error {
	container := tx.Bucket([]byte(bucket))
	if container == nil {
		return fmt.Errorf("bucket %s is missing", bucket)
	}
	if container.Get([]byte(key)) == nil {
		return domain.ErrNotFound
	}
	return container.Delete([]byte(key))
}

func listJSON[T any](tx *bolt.Tx, bucket string) ([]T, error) {
	container := tx.Bucket([]byte(bucket))
	if container == nil {
		return nil, fmt.Errorf("bucket %s is missing", bucket)
	}
	keys := make([]string, 0)
	values := make(map[string][]byte)
	err := container.ForEach(func(key, value []byte) error {
		if value == nil {
			return nil
		}
		name := string(key)
		keys = append(keys, name)
		values[name] = append([]byte(nil), value...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(keys)
	items := make([]T, 0, len(keys))
	for _, key := range keys {
		var item T
		if err := json.Unmarshal(values[key], &item); err != nil {
			return nil, fmt.Errorf("decode %s/%s: %w", bucket, key, err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Store) Count(bucket string) (int, error) {
	count := 0
	err := s.withView(func(tx *bolt.Tx) error {
		container := tx.Bucket([]byte(bucket))
		if container == nil {
			return fmt.Errorf("bucket %s is missing", bucket)
		}
		count = container.Stats().KeyN
		return nil
	})
	return count, err
}

func (s *Store) Delete(bucket, key string) error {
	return s.withUpdate(func(tx *bolt.Tx) error { return deleteValue(tx, bucket, key) })
}
