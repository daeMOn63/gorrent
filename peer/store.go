package peer

import (
	"bytes"
	"encoding/gob"
	"os"
	"time"

	"github.com/daeMOn63/gorrent/gorrent"

	bolt "go.etcd.io/bbolt"
)

var (
	gorrentBucket = []byte("gorrent")
)

// GorrentStore defines the methods needed for a Peer store
type GorrentStore interface {
	Close() error
	Save(g *gorrent.Gorrent) error
	All() ([]*gorrent.Gorrent, error)
}

type gorrentStore struct {
	db *bolt.DB
}

var _ GorrentStore = &gorrentStore{}

// NewStore creates a new peer store
func NewStore(path string, mode os.FileMode) (GorrentStore, error) {
	db, err := bolt.Open(path, mode, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	return &gorrentStore{
		db: db,
	}, nil
}

// Save save given gorrent to the store
func (s *gorrentStore) Save(g *gorrent.Gorrent) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(gorrentBucket)
		if err != nil {
			return err
		}

		data, err := s.encode(g)
		if err != nil {
			return err
		}

		err = bucket.Put(g.InfoHash().Bytes(), data)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *gorrentStore) All() ([]*gorrent.Gorrent, error) {
	var list []*gorrent.Gorrent

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(gorrentBucket)
		if bucket == nil {
			return nil
		}

		bucket.ForEach(func(k, v []byte) error {
			g, err := s.decode(v)
			if err != nil {
				return err
			}

			list = append(list, g)

			return nil
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return list, nil
}

// Close close the store
func (s *gorrentStore) Close() error {
	return s.db.Close()
}

func (s *gorrentStore) encode(g *gorrent.Gorrent) ([]byte, error) {

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(g)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *gorrentStore) decode(data []byte) (*gorrent.Gorrent, error) {
	g := &gorrent.Gorrent{}

	r := bytes.NewReader(data)
	dec := gob.NewDecoder(r)

	err := dec.Decode(g)
	if err != nil {
		return nil, err
	}

	return g, nil
}
