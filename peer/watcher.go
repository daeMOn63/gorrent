package peer

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/daeMOn63/gorrent/buffer"
	"github.com/daeMOn63/gorrent/fs"
	"github.com/daeMOn63/gorrent/gorrent"
	"github.com/daeMOn63/gorrent/tracker"
	"github.com/daeMOn63/gorrent/tracker/actions"
)

var (
	// ErrIntegrityCheckFailed is returned when the watcher fail to validate a file integrity
	ErrIntegrityCheckFailed = errors.New("integrity check failed")
)

// Watcher defines a gorrent watcher, responsible of checking the status and integrity of stored gorrent
type Watcher interface {
	Watch() error
}

type watcher struct {
	store      GorrentStore
	fs         fs.FileSystem
	fileBuffer buffer.File
	tracker    tracker.Client
}

var _ Watcher = &watcher{}

// NewWatcher creates a new gorrent watcher
func NewWatcher(store GorrentStore, fs fs.FileSystem, fileBuffer buffer.File, tracker tracker.Client) Watcher {
	return &watcher{
		store:      store,
		fs:         fs,
		fileBuffer: fileBuffer,
		tracker:    tracker,
	}
}

// Watch periodically check each stored gorrent for integrity and completion
func (w *watcher) Watch() error {

	var processedEntries []*GorrentEntry

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			entries, err := w.store.All()
			if err != nil {
				log.Println(err)
			}

			for _, entry := range entries {
				found := false
				for _, pentry := range processedEntries {
					if bytes.Equal(entry.Gorrent.InfoHash().Bytes(), pentry.Gorrent.InfoHash().Bytes()) {
						found = true
					}
				}

				if found {
					continue
				}

				go func() {
					log.Printf("Starting processing gorrent %s (%s)", entry.Name, entry.Gorrent.InfoHash().HexString())
					for {
						switch entry.Status {
						case StatusNew:
							err = w.processNew(entry)
						case StatusReady:
							err = w.processReady(entry)
						case StatusDownloading:
							err = w.processDownloading(entry)
						case StatusCompleted:
							err = w.announce(entry)
							time.Sleep(100 * time.Millisecond)
						}

						if err != nil {
							log.Println(err)
						}
					}
				}()

				processedEntries = append(processedEntries, entry)
			}
		}
	}()

	for {
		time.Sleep(1 * time.Second)
	}
}

func (w *watcher) announce(entry *GorrentEntry) error {
	nextAnnounce := entry.LastAnnounce.Add(1 * time.Second)
	if time.Now().After(nextAnnounce) {
		log.Printf("Announcing %s (%s)", entry.Name, entry.Gorrent.InfoHash().HexString())
		peers, err := w.tracker.Announce(entry.Gorrent, actions.AnnounceEventStarted, actions.AnnounceStatus{
			Downloaded: 0,
			Uploaded:   0,
		})
		if err != nil {
			return err
		}
		entry.LastAnnounce = time.Now()
		log.Printf("Got %d peers: %s", len(peers), peers)
		err = w.store.Save(entry)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *watcher) processDownloading(entry *GorrentEntry) error {

	w.announce(entry)

	time.Sleep(50 * time.Millisecond)
	// Randomly pick a missing piece
	// Ask peers for piece until getting it
	//
	return nil
}

func (w *watcher) processReady(entry *GorrentEntry) error {
	err := w.fileBuffer.Create(entry.TmpFileName(), int64(entry.Gorrent.TotalFileSize()))
	if err != nil {
		return err
	}

	entry.Status = StatusDownloading
	entry.LastAnnounce = time.Now()
	entry.LastAnnounce.Add(-1 * time.Second)

	return w.store.Save(entry)
}

func (w *watcher) processNew(entry *GorrentEntry) error {
	err := w.fs.MkdirAll(entry.Path, 0755)
	if err != nil {
		return err
	}

	isCompleted := true
	for _, gorrentFile := range entry.Gorrent.Files {
		filePath := filepath.Join(entry.Path, gorrentFile.Name)
		f, err := w.fs.Open(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("gorrent %s (%s) is not complete yet.", entry.Name, entry.Gorrent.InfoHash().HexString())
				isCompleted = false
				break
			}

			return err
		}

		log.Printf("checking integrity for file %s\n", filePath)
		if err := checkIntegrity(f, gorrentFile.Hash); err != nil {
			log.Printf("integrity check failed for file %s\n", filePath)

			entry.Status = StatusCorrupted
			w.store.Save(entry)

			return err
		}

		f.Close()
	}

	if isCompleted {
		log.Printf("gorrent %s (%s) is completed\n", entry.Name, entry.Gorrent.InfoHash().HexString())
		entry.Downloaded = entry.Gorrent.TotalFileSize()
		entry.Status = StatusCompleted
	} else {
		log.Printf("gorrent %s (%s) is ready for download\n", entry.Name, entry.Gorrent.InfoHash().HexString())
		entry.Status = StatusReady
	}

	return w.store.Save(entry)
}

func checkIntegrity(r io.Reader, expectedHash gorrent.Sha1Hash) error {
	hash := sha1.New()
	_, err := io.Copy(hash, r)
	if err != nil {
		return err
	}

	var sha1Hash gorrent.Sha1Hash
	copy(sha1Hash[:], hash.Sum(nil))

	if sha1Hash.HexString() != expectedHash.HexString() {
		return ErrIntegrityCheckFailed
	}

	return nil
}
