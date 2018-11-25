package peer

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"github.com/daeMOn63/gorrent/gorrent"

	bolt "go.etcd.io/bbolt"
)

var (
	gorrentBucket = []byte("gorrent")
)

const (
	// StatusNew is set when the gorrent has just been added to the store
	StatusNew Status = "new"
	// StatusCheck is set when the gorrent is currently being check for integrity and completion
	StatusCheck Status = "checking"
	// StatusDownloading is set when the gorrent is currently being downloaded
	StatusDownloading Status = "downloading"
	// StatusCompleted is set when the gorrent has been fully downloaded and checked
	StatusCompleted Status = "completed"
	// StatusCorrupted is set when at least one of the gorrent files failed to pass the integrity check
	StatusCorrupted Status = "corrupted"
	// StatusReady is set right after StatusNew, and before StatusDownloading, to indicate that the gorrent is ready to be downloaded
	StatusReady Status = "ready"
)

// Status defines a string type for holding gorrent status
type Status string

// GorrentStore defines the methods needed for a Peer store
type GorrentStore interface {
	Close() error
	Save(g *GorrentEntry) error
	All() ([]*GorrentEntry, error)
	Get(gorrent.Sha1Hash) (*GorrentEntry, error)
}

type gorrentStore struct {
	db *bolt.DB
}

var _ GorrentStore = &gorrentStore{}

// GorrentEntry defines data saved in the peer database
type GorrentEntry struct {
	Name            string
	Gorrent         *gorrent.Gorrent
	CreatedAt       time.Time
	Path            string
	Uploaded        uint64
	Downloaded      uint64
	Status          Status
	PeerAddrs       []gorrent.PeerAddr
	CompletedChunks []int64
}

// TmpFileName returns the temporary filename for this entry
func (g *GorrentEntry) TmpFileName() string {
	return fmt.Sprintf("%s.dat", g.Gorrent.InfoHash().HexString())
}

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
func (s *gorrentStore) Save(g *GorrentEntry) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(gorrentBucket)
		if err != nil {
			return err
		}

		data, err := s.encode(g)
		if err != nil {
			return err
		}

		err = bucket.Put(g.Gorrent.InfoHash().Bytes(), data)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *gorrentStore) Get(infoHash gorrent.Sha1Hash) (*GorrentEntry, error) {
	entry := &GorrentEntry{}

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(gorrentBucket)
		if bucket == nil {
			return nil
		}

		v := bucket.Get(infoHash.Bytes())
		if v != nil {
			var err error
			entry, err = s.decode(v)
			if err != nil {
				return err
			}

		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *gorrentStore) All() ([]*GorrentEntry, error) {
	var list []*GorrentEntry

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

func (s *gorrentStore) encode(g *GorrentEntry) ([]byte, error) {

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(g)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *gorrentStore) decode(data []byte) (*GorrentEntry, error) {
	g := &GorrentEntry{}

	r := bytes.NewReader(data)
	dec := gob.NewDecoder(r)

	err := dec.Decode(g)
	if err != nil {
		return nil, err
	}

	return g, nil
}
